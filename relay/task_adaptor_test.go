package relay

import (
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTaskAdaptorSupportedPlatforms(t *testing.T) {
	tests := []struct {
		name     string
		platform constant.TaskPlatform
		expected bool
	}{
		{
			name:     "custom channel type 8",
			platform: constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeCustom)),
			expected: true,
		},
		{
			name:     "advanced custom channel type 58",
			platform: constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeAdvancedCustom)),
			expected: true,
		},
		{
			name:     "openai channel type 1",
			platform: constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeOpenAI)),
			expected: true,
		},
		{
			name:     "sora channel type 55",
			platform: constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeSora)),
			expected: true,
		},
		{
			name:     "suno task platform",
			platform: constant.TaskPlatformSuno,
			expected: true,
		},
		{
			name:     "unknown platform",
			platform: constant.TaskPlatform("99999"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adaptor := GetTaskAdaptor(tc.platform)
			if tc.expected {
				require.NotNil(t, adaptor, "adaptor for platform %s should not be nil", tc.platform)
			} else {
				assert.Nil(t, adaptor, "adaptor for platform %s should be nil", tc.platform)
			}
		})
	}
}
