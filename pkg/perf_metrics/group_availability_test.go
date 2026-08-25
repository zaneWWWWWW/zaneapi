package perfmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGroupAvailabilityUsesCurrentBucketAndHoursWindow(t *testing.T) {
	const currentBucket int64 = 1_700_000_000
	merged := map[string]map[int64]counters{
		"default": {
			currentBucket:        {requestCount: 10, successCount: 9},
			currentBucket - 3600: {requestCount: 90, successCount: 81},
		},
		"vip": {
			currentBucket - 3600: {requestCount: 40, successCount: 40},
		},
	}

	result := buildGroupAvailability(
		[]string{"vip", "default", "empty"},
		currentBucket,
		24,
		3600,
		merged,
	)

	require.Len(t, result.Groups, 3)
	assert.Equal(t, 24, result.Hours)
	assert.Equal(t, int64(3600), result.BucketSeconds)

	byName := map[string]GroupAvailability{}
	for _, group := range result.Groups {
		byName[group.Group] = group
	}

	require.NotNil(t, byName["default"].CurrentSuccessRate)
	assert.Equal(t, 90.0, *byName["default"].CurrentSuccessRate)
	require.NotNil(t, byName["default"].HoursSuccessRate)
	assert.Equal(t, 90.0, *byName["default"].HoursSuccessRate)
	assert.Equal(t, int64(10), byName["default"].CurrentRequestCount)
	assert.Equal(t, int64(100), byName["default"].HoursRequestCount)

	assert.Nil(t, byName["vip"].CurrentSuccessRate)
	require.NotNil(t, byName["vip"].HoursSuccessRate)
	assert.Equal(t, 100.0, *byName["vip"].HoursSuccessRate)
	assert.Equal(t, int64(0), byName["vip"].CurrentRequestCount)
	assert.Equal(t, int64(40), byName["vip"].HoursRequestCount)

	assert.Nil(t, byName["empty"].CurrentSuccessRate)
	assert.Nil(t, byName["empty"].HoursSuccessRate)
	assert.Equal(t, int64(0), byName["empty"].HoursRequestCount)
}

func TestBuildGroupAvailabilitySortsByTrafficThenName(t *testing.T) {
	result := buildGroupAvailability(
		[]string{"beta", "alpha", "vip"},
		100,
		24,
		3600,
		map[string]map[int64]counters{
			"beta":  {100: {requestCount: 5, successCount: 5}},
			"alpha": {100: {requestCount: 5, successCount: 4}},
			"vip":   {100: {requestCount: 20, successCount: 20}},
		},
	)

	require.Len(t, result.Groups, 3)
	assert.Equal(t, "vip", result.Groups[0].Group)
	assert.Equal(t, "alpha", result.Groups[1].Group)
	assert.Equal(t, "beta", result.Groups[2].Group)
}
