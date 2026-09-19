package ratio_setting

import (
	"testing"

	"github.com/QuantumNous/new-api/types"
	"github.com/stretchr/testify/require"
)

func TestAgnesDefaultsPreserveSavedCustomAndZero(t *testing.T) {
	for _, tc := range []struct {
		name     string
		target   *types.RWMap[string, float64]
		defaults map[string]float64
		update   func(string) error
	}{
		{"price", modelPriceMap, defaultModelPrice, UpdateModelPriceByJSONString},
		{"input", modelRatioMap, defaultModelRatio, UpdateModelRatioByJSONString},
		{"output", completionRatioMap, defaultCompletionRatio, UpdateCompletionRatioByJSONString},
		{"cache", cacheRatioMap, defaultCacheRatio, UpdateCacheRatioByJSONString},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := tc.target.MarshalJSONString()
			t.Cleanup(func() {
				require.NoError(t, types.LoadFromJsonStringWithCallback(tc.target, previous, InvalidateExposedDataCache))
			})
			require.NoError(t, tc.update(`{"agnes-2.5-flash":0,"agnes-video-v2.0":0.123,"unrelated":7}`))
			values := tc.target.ReadAll()
			require.Equal(t, 0.0, values["agnes-2.5-flash"])
			require.Equal(t, 0.123, values["agnes-video-v2.0"])
			require.Equal(t, 7.0, values["unrelated"])
			for name, value := range tc.defaults {
				if len(name) >= 6 && name[:6] == "agnes-" && name != "agnes-2.5-flash" && name != "agnes-video-v2.0" {
					require.Equal(t, value, values[name], name)
				}
			}
		})
	}
}

func TestAgnesStandardPriceDefaults(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		input, output, cache float64
	}{
		{"agnes-2.5-flash", 0.05, 0.15, 0.005}, {"agnes-3.0-flash", 0.05, 0.15, 0.005}, {"agnes-2.5-pro-beta", 0.10, 0.30, 0.01}, {"agnes-2.5-pro", 0.45, 0.90, 0.045},
	} {
		input := defaultModelRatio[tc.name] * 2
		require.InDelta(t, tc.input, input, 1e-9)
		require.InDelta(t, tc.output, input*defaultCompletionRatio[tc.name], 1e-9)
		require.InDelta(t, tc.cache, input*defaultCacheRatio[tc.name], 1e-9)
	}
	for _, name := range []string{"agnes-image-2.0-flash", "agnes-image-2.1-flash", "agnes-image-2.5-flash"} {
		require.Equal(t, 0.01, defaultModelPrice[name])
	}
	require.Equal(t, 0.025, defaultModelPrice["agnes-video-v2.0"])
	require.Equal(t, 0.125, defaultModelPrice["agnes-video-2.5"])
	require.Equal(t, 0.125, defaultModelPrice["agnes-video-2.5-flash"])
}
