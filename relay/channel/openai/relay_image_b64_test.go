package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/stretchr/testify/assert"
)

func TestWantsOpenAIImageB64JSONPlayground(t *testing.T) {
	t.Parallel()

	assert.True(t, wantsOpenAIImageB64JSON(&relaycommon.RelayInfo{
		IsPlayground: true,
		RelayMode:    relayconstant.RelayModeImagesEdits,
	}))
	assert.True(t, wantsOpenAIImageB64JSON(&relaycommon.RelayInfo{
		Request: &dto.ImageRequest{ResponseFormat: "b64_json"},
	}))
	assert.False(t, wantsOpenAIImageB64JSON(&relaycommon.RelayInfo{
		Request: &dto.ImageRequest{ResponseFormat: "url"},
	}))
}
