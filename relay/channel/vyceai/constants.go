package vyceai

const (
	ChannelName      = "vyceai"
	UpstreamModel    = "default"
	UpstreamSize     = "4096x4096"
	ImageStreamPath  = "/v1/images/stream"
	maxSSEEventBytes = 64 << 20
)

var ModelList = []string{
	"vyce-image-1x1",
	"vyce-image-16x9",
	"vyce-image-9x16",
	"vyce-image-2x3",
	"vyce-image-3x2",
	"vyce-image-4x3",
	"你妈-1x1",
	"你妈-16x9",
	"你妈-9x16",
	"你妈-2x3",
	"你妈-3x2",
	"你妈-4x3",
}

var modelAspectRatios = map[string]string{
	"vyce-image-1x1":  "1:1",
	"vyce-image-16x9": "16:9",
	"vyce-image-9x16": "9:16",
	"vyce-image-2x3":  "2:3",
	"vyce-image-3x2":  "3:2",
	"vyce-image-4x3":  "4:3",
	"你妈-1x1":          "1:1",
	"你妈-16x9":         "16:9",
	"你妈-9x16":         "9:16",
	"你妈-2x3":          "2:3",
	"你妈-3x2":          "3:2",
	"你妈-4x3":          "4:3",
}

func aspectRatioForModel(model string) (string, bool) {
	ratio, ok := modelAspectRatios[model]
	return ratio, ok
}
