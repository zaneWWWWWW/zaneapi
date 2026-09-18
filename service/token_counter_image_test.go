package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 生图 prompt 只用本地估算：tiktoken 对超长文本开销超线性（实测 100KB 约 4.6s），
// 而 prompt token 数不参与生图定价（按次价/模型倍率都取 MaxTokens），只用于日志展示。
func TestEstimateRequestTokenUsesLocalEstimatorForImageRelay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	prevCountToken := constant.CountToken
	constant.CountToken = true
	defer func() { constant.CountToken = prevCountToken }()

	const (
		model = "gpt-image-2"
		// 中英混合，保证本地估算与 tiktoken 结果不同，可区分实际走的分支。
		prompt = "赛博朋克霓虹都市夜景，reflection on wet asphalt, cinematic lighting"
	)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	common.SetContextKey(ctx, constant.ContextKeyOriginalModel, model)

	tokens, err := EstimateRequestToken(ctx, &types.TokenCountMeta{
		CombineText: prompt,
		MaxTokens:   1584,
	}, &relaycommon.RelayInfo{RelayFormat: types.RelayFormatOpenAIImage})
	require.NoError(t, err)

	// 该 fixture 下 tiktoken 与本地估算结果不同，若回退到 CountTextToken 这里会失败。
	assert.Equal(t, EstimateTokenByModel(model, prompt), tokens,
		"生图 prompt 应使用本地估算，而不是 tiktoken 分词")
}
