package constant

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPath2RelayModeVideoEndpoints(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{path: "/v1/video/generations", want: RelayModeVideoSubmit},
		{path: "/v1/videos/generations", want: RelayModeVideoSubmit},
		{path: "/v1/videos", want: RelayModeVideoSubmit},
		{path: "/v1/videos/video_123/remix", want: RelayModeVideoSubmit},
		{path: "/v1/video/generations/task_123", want: RelayModeVideoFetchByID},
		{path: "/v1/videos/task_123", want: RelayModeVideoFetchByID},
		{path: "/pg/videos", want: RelayModeVideoSubmit},
		{path: "/pg/videos/generations", want: RelayModeVideoSubmit},
		{path: "/pg/videos/task_123", want: RelayModeVideoFetchByID},
		{path: "/pg/images/generations", want: RelayModeImagesGenerations},
		{path: "/pg/images/edits", want: RelayModeImagesEdits},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			require.Equal(t, tt.want, Path2RelayMode(tt.path))
		})
	}
}
