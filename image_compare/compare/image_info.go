package compare

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/corona10/goimagehash"
)

type ImageInfo struct {
	Error string

	Width    int
	Height   int
	Format   string
	Quality  int
	FrameNum int

	AverageHash    *goimagehash.ImageHash
	DifferenceHash *goimagehash.ImageHash
	PerceptionHash *goimagehash.ImageHash
}

func ParseImageInfo(data []byte) *ImageInfo {
	info := &ImageInfo{
		AverageHash:    goimagehash.NewImageHash(100, goimagehash.Unknown),
		DifferenceHash: goimagehash.NewImageHash(100, goimagehash.Unknown),
		PerceptionHash: goimagehash.NewImageHash(100, goimagehash.Unknown),
	}
	if !ParseImageBaseInfo(info, data) || !ParseImageHash(info, data) {
		return info
	}
	return info
}

func ParseImageBaseInfo(info *ImageInfo, data []byte) bool {
	r := bytes.NewReader(data)
	img, format, err := image.DecodeConfig(r)
	if err != nil {
		info.Error = err.Error()
		return false
	}
	info.Width = img.Width
	info.Height = img.Height
	info.Format = format
	// TODO 获取图片质量和GIF帧数
	return true
}

func ParseImageHash(info *ImageInfo, data []byte) bool {
	r := bytes.NewReader(data)
	img, _, err := image.Decode(r)
	if err != nil {
		info.Error = err.Error()
		return false
	}
	info.AverageHash, _ = goimagehash.AverageHash(img)
	info.DifferenceHash, _ = goimagehash.DifferenceHash(img)
	info.PerceptionHash, _ = goimagehash.PerceptionHash(img)
	return true
}
func Compare(ori *ImageInfo, replay *ImageInfo, result map[string]Diff) {
	if ori.Error != replay.Error {
		result[IMAGE_ERROR] = &StringDiff{ori.Error, replay.Error}
	}
	if ori.Width != replay.Width {
		result[WIDTH] = &IntDiff{ori.Width, replay.Width}
	}
	if ori.Height != replay.Height {
		result[HEIGHT] = &IntDiff{ori.Height, replay.Height}
	}
	if ori.Quality != replay.Quality {
		result[QUALITY] = &IntDiff{ori.Quality, replay.Quality}
	}
	if ori.Format != replay.Format {
		result[FORMAT] = &StringDiff{ori.Format, replay.Format}
	}
	if ori.FrameNum != replay.FrameNum {
		result[FRAME_NUM] = &IntDiff{ori.FrameNum, replay.FrameNum}
	}
	if dis, _ := ori.AverageHash.Distance(replay.AverageHash); dis != 0 {
		result[AVERAGE_HASH] = &IntDiff{dis, 0}
	}
	if dis, _ := ori.DifferenceHash.Distance(replay.DifferenceHash); dis != 0 {
		result[DIFFERENCE_HASH] = &IntDiff{dis, 0}
	}
	if dis, _ := ori.PerceptionHash.Distance(replay.PerceptionHash); dis != 0 {
		result[PERCEPTION_HASH] = &IntDiff{dis, 0}
	}
}
