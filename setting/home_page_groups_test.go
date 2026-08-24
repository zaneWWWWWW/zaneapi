package setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHomePageDisplayedGroups(t *testing.T) {
	assert.Empty(t, ParseHomePageDisplayedGroups(""))
	assert.Empty(t, ParseHomePageDisplayedGroups("not-json"))
	assert.Equal(t, []string{"default", "vip"}, ParseHomePageDisplayedGroups(`["default","vip","default"," "]`))
}

func TestBuildHomePageGroupRatiosSkipsUnknownGroups(t *testing.T) {
	items := BuildHomePageGroupRatios(`["default","missing","vip"]`)
	require.Len(t, items, 2)
	assert.Equal(t, "default", items[0].Name)
	assert.Equal(t, 1.0, items[0].Ratio)
	assert.Equal(t, "vip", items[1].Name)
}

func TestBuildHomePageGroupRatiosEmptyWhenUnset(t *testing.T) {
	assert.Empty(t, BuildHomePageGroupRatios(""))
	assert.Empty(t, BuildHomePageGroupRatios("[]"))
	assert.NotNil(t, BuildHomePageGroupRatios("[]"))
}
