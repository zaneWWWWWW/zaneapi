package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
)

func TestShouldPassThroughImageBody(t *testing.T) {
	t.Parallel()

	assert.False(t, shouldPassThroughImageBody(nil))
	assert.False(t, shouldPassThroughImageBody(&relaycommon.RelayInfo{}))

	gemini := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType: constant.ChannelTypeGemini,
			ChannelSetting: dto.ChannelSettings{
				PassThroughBodyEnabled: true,
			},
		},
	}
	assert.False(t, shouldPassThroughImageBody(gemini))

	openai := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType: constant.ChannelTypeOpenAI,
			ChannelSetting: dto.ChannelSettings{
				PassThroughBodyEnabled: true,
			},
		},
	}
	assert.True(t, shouldPassThroughImageBody(openai))
}
