package perfmetrics

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisGroupBucketsShareActiveMetricsAcrossNodes(t *testing.T) {
	server := miniredis.RunT(t)
	oldRedisEnabled := common.RedisEnabled
	oldRedisClient := common.RDB
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	require.NoError(t, client.Ping(t.Context()).Err())
	common.RedisEnabled = true
	common.RDB = client
	t.Cleanup(func() {
		_ = client.Close()
		common.RedisEnabled = oldRedisEnabled
		common.RDB = oldRedisClient
	})

	key := bucketKey{model: "gpt-test", group: "vip", bucketTs: 1_700_000_000}
	recordRedis(key, Sample{
		Success:      true,
		LatencyMs:    200,
		OutputTokens: 20,
		GenerationMs: 100,
	})

	merged := map[string]map[int64]counters{}
	mergeRedisGroupActiveBuckets(merged, key.bucketTs, []string{"vip"})

	value := merged["vip"][key.bucketTs]
	assert.Equal(t, int64(1), value.requestCount)
	assert.Equal(t, int64(1), value.successCount)
	assert.Equal(t, int64(200), value.totalLatencyMs)
}
