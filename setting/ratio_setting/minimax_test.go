package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinimaxDefaultModelPrices(t *testing.T) {
	expectedPrices := map[string]float64{
		"minimax-h3":       0.1,
		"minimax-h3-turbo": 0.05,
		"minimax-video":    0.1,
		"minimax/video-01": 0.1,
		"h3":               0.1,
	}

	for model, expected := range expectedPrices {
		price, ok := defaultModelPrice[model]
		require.True(t, ok, "model %s should exist in defaultModelPrice", model)
		assert.Equal(t, expected, price, "model %s should have expected price", model)
	}
}
