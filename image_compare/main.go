/*
This middleware made for auth system that randomly generate access tokens, which used later for accessing secure content. Since there is no pre-defined token value, naive approach without middleware (or if middleware use only request payloads) will fail, because replayed server have own tokens, not synced with origin. To fix this, our middleware should take in account responses of replayed and origin server, store `originalToken -> replayedToken` aliases and rewrite all requests using this token to use replayed alias. See `middleware_test.go#TestTokenMiddleware` test for examples of using this middleware.

How middleware works:

                   Original request      +--------------+
+-------------+----------STDIN---------->+              |
|  Gor input  |                          |  Middleware  |
+-------------+----------STDIN---------->+              |
                   Original response     +------+---+---+
                                                |   ^
+-------------+    Modified request             v   |
| Gor output  +<---------STDOUT-----------------+   |
+-----+-------+                                     |
      |                                             |
      |            Replayed response                |
      +------------------STDIN----------------->----+

图片对比程序

1. 抓取线上包
```
./gor --input-raw :8080  -input-raw-track-response --output-file requests.gor
```
添加`-output-http-track-response`， 是为了保存HTTP的返回值

2. 执行对比
```
./goreplay --input-file ./requests_1.gor  --middleware "examples/middleware/image_compare" --input-raw-track-response  --output-http "http://****"
```
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/buger/goreplay/image_compare/compare"
	_ "github.com/buger/goreplay/proto"
)

const NEWLINE byte = '\n'

var num int
var bufSize int
var chanSize int

// requestID -> request
var reqMap = make(map[string][]byte)
var reqMapMutex sync.Mutex

// requestID -> response
var respMap = make(map[string][]byte)
var respMapMutex sync.Mutex

func main() {
	flag.IntVar(&num, "n", 1, "compare worker num")
	flag.IntVar(&bufSize, "size", 2*20*1024*1024, "read buffer size")
	flag.IntVar(&chanSize, "ch", 32, "channel size")
	flag.Parse()

	ch := make(chan []byte, chanSize)
	for i := 0; i < num; i++ {
		go worker(ch)
	}

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, bufSize)
	scanner.Buffer(buf, bufSize)
	for scanner.Scan() {
		encoded := scanner.Bytes()
		data := make([]byte, len(encoded))
		copy(data, encoded)
		select {
		case ch <- data:
		}
	}
	Debug("stop ...", scanner.Err())
	close(ch)
	for {
		Debug("wait process all request")
		if len(ch) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	// FIXME wait request
}

func worker(ch chan []byte) {
	var data []byte
	var ok bool
	for {
		select {
		case data, ok = <-ch:
			if !ok {
				return
			}
		}
		decodeData := make([]byte, len(data)/2)
		hex.Decode(decodeData, data)
		process(data, decodeData)
	}
}

// 处理请求
// encodeData 为使用hex编码后的数据
// decodeData 为已经解码之后的数据
func process(encodeData []byte, decodeData []byte) {
	// First byte indicate payload type, possible values:
	//  1 - Request
	//  2 - Response
	//  3 - ReplayedResponse

	payloadType := decodeData[0]
	headerSize := bytes.IndexByte(decodeData, '\n') + 1
	header := decodeData[:headerSize-1]

	// Header contains space separated values of: request type, request id, and request start time (or round-trip time for responses)
	meta := bytes.Split(header, []byte(" "))
	// For each request you should receive 3 payloads (request, response, replayed response) with same request id
	reqID := string(meta[1])
	payload := decodeData[headerSize:]

	Debug("process: ", string(header))

	switch payloadType {
	case '1':
		// TODO 更改鉴权的token
		encodeData = append(encodeData, NEWLINE)
		os.Stdout.Write(encodeData)
		putRequest(reqID, payload)
	case '2':
		resp := getOrSet(reqID, payload)
		if resp != nil {
			req := getRequest(reqID)
			Compare(reqID, req, payload, resp)
		}
	case '3':
		resp := getOrSet(reqID, payload)
		if resp != nil {
			req := getRequest(reqID)
			Compare(reqID, req, resp, payload)
		}
	}
}
func putRequest(reqID string, payload []byte) {
	reqMapMutex.Lock()
	defer reqMapMutex.Unlock()
	reqMap[reqID] = payload
}

func getRequest(reqID string) []byte {
	reqMapMutex.Lock()
	defer reqMapMutex.Unlock()
	if val, ok := reqMap[reqID]; ok {
		return val
	}
	return nil
}

func getOrSet(reqID string, payload []byte) []byte {
	respMapMutex.Lock()
	defer respMapMutex.Unlock()
	if val, ok := respMap[reqID]; ok {
		delete(respMap, reqID)
		return val
	}
	respMap[reqID] = payload
	return nil
}

func Compare(reqID string, req []byte, oriResp []byte, replayResp []byte) {
	result := make(map[string]compare.Diff)
	compare.CompareHttp(oriResp, replayResp, result)
	if len(result) != 0 {
		compare.CompareImage(oriResp, replayResp, result)
	}
	if len(result) == 0 {
		Debug("%s request is same", reqID)
	} else {
		Debug("%s request is diff %v", reqID, result)
	}
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'
	return dst
}

func Debug(args ...interface{}) {
	if os.Getenv("GOR_TEST") != "" { // if we are not testing
		fmt.Fprint(os.Stderr, "[DEBUG][IMAGE_COMP] ")
		fmt.Fprintln(os.Stderr, args...)
	}
}
