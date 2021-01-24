package compare

/**
 * 对比HTTP内容的差异
 * 1. HTTP状态码
 * 2. HTTP头部
 * 3. 文件大小
 */

import (
	"bytes"
	"crypto/md5"

	"github.com/buger/goreplay/proto"
)

func CompareHttp(oriResp []byte, replayResp []byte, result map[string]Diff) {
	oriStatus := proto.Status(oriResp)
	replayStatus := proto.Status(replayResp)
	if bytes.Compare(oriStatus, replayStatus) != 0 {
		result[STATUS] = &BytesDiff{Ori: oriStatus, Replay: replayStatus}
	}

	for _, key := range COMPARE_HEADERS {
		ori := proto.Header(oriResp, key)
		replay := proto.Header(replayResp, key)
		if bytes.Compare(ori, replay) != 0 {
			result[string(key)] = &BytesDiff{ori, replay}
		}
	}

	oriSum := md5.Sum(proto.Body(oriResp))
	replaySum := md5.Sum(proto.Body(oriResp))
	if bytes.Compare(oriSum[:], replaySum[:]) != 0 {
		result[MD5] = &BytesDiff{Ori: oriSum[:], Replay: replaySum[:]}
	}
}
