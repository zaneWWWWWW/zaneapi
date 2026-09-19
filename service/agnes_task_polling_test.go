package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

type agnesPollingFixture struct {
	status         string
	httpStatus     int
	transportError bool
	started        chan struct{}
	release        chan struct{}
	calls          atomic.Int32
	settlements    atomic.Int32
	params         chan map[string]any
}

func (*agnesPollingFixture) Init(*relaycommon.RelayInfo) {}
func (a *agnesPollingFixture) FetchTask(_ string, key string, body map[string]any, _ string) (*http.Response, error) {
	a.calls.Add(1)
	if a.params != nil {
		body["key"] = key
		a.params <- body
	}
	if a.started != nil {
		a.started <- struct{}{}
		<-a.release
	}
	if a.transportError {
		return nil, fmt.Errorf("temporary transport error")
	}
	status := a.httpStatus
	if status == 0 {
		status = 200
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(`{"status":"fixture"}`))}, nil
}
func (a *agnesPollingFixture) ParseTaskResult([]byte) (*relaycommon.TaskInfo, error) {
	return &relaycommon.TaskInfo{Status: a.status, Reason: "upstream failed", Url: "https://example.com/result.mp4", TotalTokens: 999999}, nil
}
func (a *agnesPollingFixture) AdjustBillingOnComplete(*model.Task, *relaycommon.TaskInfo) int {
	a.settlements.Add(1)
	return 0
}

func seedAgnesPollingTask(t *testing.T, source string) (*model.Task, *model.Channel) {
	t.Helper()
	truncate(t)
	seedUser(t, 1, 100000)
	seedToken(t, 1, 1, "test-token", 100000)
	seedChannel(t, 1)
	if source == BillingSourceSubscription {
		seedSubscription(t, 1, 1, 200000, 62500)
	}
	task := makeTask(1, 1, 62500, 1, source, 1)
	task.TaskID = "public-agnes-task"
	task.Platform = constant.TaskPlatform(fmt.Sprint(constant.ChannelTypeAgnesAI))
	task.Properties.UpstreamModelName = "agnes-video-2.5"
	task.PrivateData.UpstreamTaskID = "upstream-task"
	task.PrivateData.UpstreamVideoID = "upstream-video"
	task.PrivateData.Key = "saved-channel-key"
	task.PrivateData.BillingContext.PerCallBilling = true
	task.PrivateData.BillingContext.ModelPrice = 0.125
	require.NoError(t, model.DB.Create(task).Error)
	return task, &model.Channel{Id: 1, Type: constant.ChannelTypeAgnesAI, Key: "rotated-key"}
}

func reloadAgnesTask(t *testing.T, id int64) *model.Task {
	t.Helper()
	var task model.Task
	require.NoError(t, model.DB.First(&task, id).Error)
	return &task
}

func TestAgnesConcurrentPollingRefundsOnceAfterReload(t *testing.T) {
	for _, source := range []string{BillingSourceWallet, BillingSourceSubscription} {
		t.Run(source, func(t *testing.T) {
			task, ch := seedAgnesPollingTask(t, source)
			// Independent database loads reproduce a restart and overlapping
			// foreground/background requests with the same non-terminal status.
			first, second := reloadAgnesTask(t, task.ID), reloadAgnesTask(t, task.ID)
			require.Equal(t, "upstream-video", first.PrivateData.UpstreamVideoID)
			a := &agnesPollingFixture{status: model.TaskStatusFailure, started: make(chan struct{}, 2), release: make(chan struct{}), params: make(chan map[string]any, 2)}
			var wg sync.WaitGroup
			errs := make(chan error, 2)
			for _, copy := range []*model.Task{first, second} {
				wg.Add(1)
				go func(task *model.Task) { defer wg.Done(); errs <- RefreshVideoTask(context.Background(), a, ch, task) }(copy)
			}
			for i := 0; i < 2; i++ {
				select {
				case <-a.started:
				case <-time.After(5 * time.Second):
					t.Fatal("concurrent poll did not start")
				}
			}
			close(a.release)
			wg.Wait()
			close(errs)
			for err := range errs {
				require.NoError(t, err)
			}
			for i := 0; i < 2; i++ {
				params := <-a.params
				require.Equal(t, "upstream-video", params["video_id"])
				require.Equal(t, "agnes-video-2.5", params["model_name"])
				require.Equal(t, "saved-channel-key", params["key"])
			}
			saved := reloadAgnesTask(t, task.ID)
			require.Equal(t, model.TaskStatus(model.TaskStatusFailure), saved.Status)
			require.Equal(t, task.PrivateData.BillingContext, saved.PrivateData.BillingContext)
			require.NoError(t, RefreshVideoTask(context.Background(), a, ch, saved))
			require.Equal(t, int32(2), a.calls.Load())
			require.Equal(t, int64(1), countLogs(t))
			require.Equal(t, 162500, getTokenRemainQuota(t, 1))
			if source == BillingSourceWallet {
				require.Equal(t, 162500, getUserQuota(t, 1))
			} else {
				require.Equal(t, int64(0), getSubscriptionUsed(t, 1))
				require.Equal(t, 100000, getUserQuota(t, 1))
			}
		})
	}
}

func TestAgnesPollingKeepsFixedAndHistoricalBilling(t *testing.T) {
	for _, historical := range []bool{false, true} {
		t.Run(fmt.Sprint("historical=", historical), func(t *testing.T) {
			task, ch := seedAgnesPollingTask(t, BillingSourceWallet)
			if historical {
				task.PrivateData.BillingContext.OtherRatios = map[string]float64{"seconds": 10}
				task.PrivateData.BillingContext.ModelPrice = 0.005
				require.NoError(t, task.Update())
			}
			a := &agnesPollingFixture{status: model.TaskStatusSuccess}
			reloaded := reloadAgnesTask(t, task.ID)
			require.NoError(t, RefreshVideoTask(context.Background(), a, ch, reloaded))
			saved := reloadAgnesTask(t, task.ID)
			require.Equal(t, task.Quota, saved.Quota)
			require.Equal(t, task.PrivateData.BillingContext, saved.PrivateData.BillingContext)
			require.Equal(t, "https://example.com/result.mp4", saved.GetResultURL())
			require.NoError(t, RefreshVideoTask(context.Background(), a, ch, saved))
			require.Zero(t, a.settlements.Load())
			require.Equal(t, 100000, getUserQuota(t, 1))
			require.Zero(t, countLogs(t))
		})
	}
}

func TestAgnesPollingTemporaryErrorsPreserveTask(t *testing.T) {
	for _, code := range []int{0, 429, 500, 503} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			task, ch := seedAgnesPollingTask(t, BillingSourceWallet)
			a := &agnesPollingFixture{httpStatus: code, transportError: code == 0, status: model.TaskStatusFailure}
			require.Error(t, RefreshVideoTask(context.Background(), a, ch, task))
			require.Equal(t, model.TaskStatus(model.TaskStatusInProgress), reloadAgnesTask(t, task.ID).Status)
			require.Equal(t, 100000, getUserQuota(t, 1))
			require.Zero(t, countLogs(t))
		})
	}
}

func TestAgnesTimeoutRefundsOnce(t *testing.T) {
	task, _ := seedAgnesPollingTask(t, BillingSourceWallet)
	previous := constant.TaskTimeoutMinutes
	constant.TaskTimeoutMinutes = 1
	t.Cleanup(func() { constant.TaskTimeoutMinutes = previous })
	task.SubmitTime = time.Now().Add(-2 * time.Minute).Unix()
	require.NoError(t, task.Update())
	sweepTimedOutTasks(context.Background())
	sweepTimedOutTasks(context.Background())
	require.Equal(t, model.TaskStatus(model.TaskStatusFailure), reloadAgnesTask(t, task.ID).Status)
	require.Equal(t, 162500, getUserQuota(t, 1))
	require.Equal(t, int64(1), countLogs(t))
	public, err := common.Marshal(task)
	require.NoError(t, err)
	require.NotContains(t, string(public), "upstream-video")
}
