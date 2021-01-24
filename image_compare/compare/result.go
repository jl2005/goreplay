package compare

const (
	STATUS = "status"
	HEADER = "header"
	SIZE   = "size"
	MD5    = "md5"

	IMAGE_ERROR     = "image_error"
	WIDTH           = "width"
	HEIGHT          = "height"
	FORMAT          = "format"
	QUALITY         = "quality"
	FRAME_NUM       = "frame_num"
	AVERAGE_HASH    = "average_hash"
	DIFFERENCE_HASH = "difference_hash"
	PERCEPTION_HASH = "perception_hash"
)

var COMPARE_HEADERS = [][]byte{
	[]byte("Content-Length"),
	[]byte("Content-Type"),
}

type Diff interface {
	Result() (interface{}, interface{})
}

type StringDiff struct {
	Ori    string
	Replay string
}

func (diff *StringDiff) Result() (interface{}, interface{}) {
	return diff.Ori, diff.Replay
}

type IntDiff struct {
	Ori    int
	Replay int
}

func (diff *IntDiff) Result() (interface{}, interface{}) {
	return diff.Ori, diff.Replay
}

type BytesDiff struct {
	Ori    []byte
	Replay []byte
}

func (diff *BytesDiff) Result() (interface{}, interface{}) {
	return diff.Ori, diff.Replay
}
