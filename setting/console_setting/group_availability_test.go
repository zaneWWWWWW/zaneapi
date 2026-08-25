package console_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGroupAvailabilityGroupsAcceptsEmptyAndUniqueNames(t *testing.T) {
	require.NoError(t, ValidateConsoleSettings("", "GroupAvailabilityGroups"))
	require.NoError(t, ValidateConsoleSettings("[]", "GroupAvailabilityGroups"))
	require.NoError(t, ValidateConsoleSettings(`["default","vip","auto"]`, "GroupAvailabilityGroups"))
}

func TestValidateGroupAvailabilityGroupsRejectsInvalidPayloads(t *testing.T) {
	require.Error(t, ValidateConsoleSettings("{", "GroupAvailabilityGroups"))
	require.Error(t, ValidateConsoleSettings(`["default","default"]`, "GroupAvailabilityGroups"))
	require.Error(t, ValidateConsoleSettings(`["bad group"]`, "GroupAvailabilityGroups"))
	require.Error(t, ValidateConsoleSettings(`["<script>"]`, "GroupAvailabilityGroups"))
}

func TestGetGroupAvailabilityGroupsDropsBlanksAndDuplicates(t *testing.T) {
	original := GetConsoleSetting().GroupAvailabilityGroups
	t.Cleanup(func() {
		GetConsoleSetting().GroupAvailabilityGroups = original
	})

	GetConsoleSetting().GroupAvailabilityGroups = `[" default ","vip","default","", "auto"]`
	assert.Equal(t, []string{"default", "vip", "auto"}, GetGroupAvailabilityGroups())

	GetConsoleSetting().GroupAvailabilityGroups = ""
	assert.Equal(t, []string{}, GetGroupAvailabilityGroups())
}
