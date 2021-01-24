package compare

/**
* 对比图片的差异
* 1. 图片长
* 2. 图片宽
* 3. 图片格式
* 4. 图片的质量信息
* 5. 图片hash
* 6. 对动图来说，对比其动图帧数
 */

import (
	"github.com/buger/goreplay/proto"
)

func CompareImage(oriResp []byte, replayResp []byte, result map[string]Diff) {
	oriBody := proto.Body(oriResp)
	replayBody := proto.Body(replayResp)

	oriImageInfo := ParseImageInfo(oriBody)
	replayImageInfo := ParseImageInfo(replayBody)

	Compare(oriImageInfo, replayImageInfo, result)
}
