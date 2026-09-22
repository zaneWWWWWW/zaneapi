package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaygroundRequestedGroup(t *testing.T) {
	t.Parallel()

	t.Run("uses body group on playground image routes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "vip", playgroundRequestedGroup("/pg/images/generations", "vip", "default"))
	})

	t.Run("falls back to New-Api-Group header when body group is empty", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "vip", playgroundRequestedGroup("/pg/images/edits", "", "vip"))
	})

	t.Run("keeps chat playground body group", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "vip", playgroundRequestedGroup("/pg/chat/completions", "vip", ""))
	})

	t.Run("ignores group on official API paths", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", playgroundRequestedGroup("/v1/images/generations", "vip", "vip"))
	})
}

func TestPlaygroundGroupAccessible(t *testing.T) {
	original := setting.UserUsableGroups2JSONString()
	t.Cleanup(func() {
		require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(original))
	})
	require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(
		`{"逆向香蕉2生图分组":"banana","生图: gpt-image(adobe)":"adobe"}`,
	))

	gin.SetMode(gin.TestMode)
	contextWithGroups := func(userGroup, usingGroup string) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/generations", nil)
		if userGroup != "" {
			common.SetContextKey(c, constant.ContextKeyUserGroup, userGroup)
		}
		if usingGroup != "" {
			common.SetContextKey(c, constant.ContextKeyUsingGroup, usingGroup)
		}
		return c
	}

	assert.True(t, playgroundGroupAccessible(contextWithGroups("default", "default"), "default", "default"))
	assert.True(t, playgroundGroupAccessible(contextWithGroups("default", "default"), "default", "逆向香蕉2生图分组"))
	assert.True(t, playgroundGroupAccessible(contextWithGroups("default", "default"), "default", "生图: gpt-image(adobe)"))
	assert.False(t, playgroundGroupAccessible(contextWithGroups("default", "default"), "default", "vip"))
	assert.False(t, playgroundGroupAccessible(contextWithGroups("", ""), "", "default"))
	assert.True(t, playgroundGroupAccessible(contextWithGroups("", ""), "", "逆向香蕉2生图分组"))
}

func TestGetModelRequestCopiesPlaygroundImageJSONGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	body := `{"model":"gemini-3.1-flash-image-preview-time","prompt":"a cat","group":"vip"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/generations", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	req, shouldSelect, err := getModelRequest(c)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.True(t, shouldSelect)
	assert.Equal(t, "gemini-3.1-flash-image-preview-time", req.Model)
	assert.Equal(t, "vip", req.Group)
}
