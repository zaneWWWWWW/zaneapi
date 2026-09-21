package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
