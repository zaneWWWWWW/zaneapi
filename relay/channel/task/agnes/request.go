package agnes

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func prepareRequest(req map[string]any, info *relaycommon.RelayInfo) (map[string]any, error) {
	modelName := getString(req, "model")
	if info != nil && info.ChannelMeta != nil && info.UpstreamModelName != "" {
		modelName = info.UpstreamModelName
	}
	for _, key := range []string{"model", "prompt"} {
		if value, ok := req[key].(string); !ok || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s must be a non-empty string", key)
		}
	}
	if modelName != ModelVideo25 && modelName != ModelVideo25Flash {
		if err := validateFrameOptions(req); err != nil {
			return nil, err
		}
		if value, ok := req["n"]; ok {
			if n, valid := numberToInt(value); !valid || n != 1 {
				return nil, fmt.Errorf("n must be 1")
			}
		}
		if _, ok := req["seconds"]; ok {
			return nil, fmt.Errorf("Agnes video v2.0 uses num_frames and frame_rate, not seconds")
		}
		payload := buildUpstreamRequest(req)
		payload["model"] = modelName
		return payload, nil
	}

	// SDK extra_body is normally flattened by the SDK. Also accept the nested
	// spelling from direct HTTP clients, with explicit top-level fields winning.
	merged := make(map[string]any)
	if extra, exists := req["extra_body"]; exists && extra != nil {
		fields, ok := extra.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("extra_body must be an object")
		}
		for key, value := range fields {
			merged[key] = value
		}
	}
	for key, value := range req {
		if key != "extra_body" {
			merged[key] = value
		}
	}
	for _, key := range []string{"image", "width", "height", "fps", "frame_rate", "num_frames", "quality", "num_inference_steps", "negative_prompt", "video_url", "video_path", "video_reference", "input_reference", "reference_url"} {
		if _, exists := merged[key]; exists {
			return nil, fmt.Errorf("%s is not supported by %s", key, modelName)
		}
	}
	mode, _ := merged["mode"].(string)
	if mode != "text" && mode != "keyframe" && mode != "reference" {
		return nil, fmt.Errorf("mode must be text, keyframe, or reference")
	}
	if _, exists := merged["seconds"]; !exists {
		merged["seconds"] = "5"
	}
	seconds, ok := merged["seconds"].(string)
	duration, err := strconv.Atoi(seconds)
	if !ok || err != nil || duration < 4 || duration > 12 {
		return nil, fmt.Errorf("seconds must be a string integer from 4 to 12")
	}
	if _, exists := merged["size"]; !exists {
		merged["size"] = "720P"
	}
	size, _ := merged["size"].(string)
	if modelName == ModelVideo25Flash && size != "720P" {
		return nil, fmt.Errorf("size must be 720P")
	}
	if size != "720P" && size != "1080P" && size != "1K" && size != "2K" {
		return nil, fmt.Errorf("size must be 720P, 1080P, 1K, or 2K")
	}
	if ratio, exists := merged["aspect_ratio"]; exists {
		switch ratio {
		case "21:9", "16:9", "4:3", "1:1", "3:4", "9:16":
		default:
			return nil, fmt.Errorf("unsupported aspect_ratio")
		}
	}
	if value, exists := merged["n"]; exists {
		if n, valid := jsonInteger(value); !valid || n != 1 {
			return nil, fmt.Errorf("n must be 1")
		}
	}
	if value, exists := merged["seed"]; exists {
		if _, valid := jsonInteger(value); !valid {
			return nil, fmt.Errorf("seed must be an integer")
		}
	}
	imageLimit := 8
	if modelName == ModelVideo25Flash {
		imageLimit = 5
	}
	images, err := validateURLArray(merged, "images", imageLimit)
	if err != nil {
		return nil, err
	}
	audios, err := validateURLArray(merged, "audios", 3)
	if err != nil {
		return nil, err
	}
	videos, err := validateVideos(merged, modelName == ModelVideo25Flash)
	if err != nil {
		return nil, err
	}
	frames := 0
	for _, key := range []string{"first_frame", "last_frame"} {
		if value, exists := merged[key]; exists && value != nil {
			if !publicMediaURL(value) {
				return nil, fmt.Errorf("%s must be an HTTP image URL", key)
			}
			frames++
		}
	}
	media := images + audios + videos
	if media > 12 {
		return nil, fmt.Errorf("reference media count must not exceed 12")
	}
	switch mode {
	case "text":
		if frames+media != 0 {
			return nil, fmt.Errorf("text mode does not accept reference media or frames")
		}
	case "keyframe":
		if frames == 0 || media != 0 {
			return nil, fmt.Errorf("keyframe requires first_frame or last_frame and does not accept reference media")
		}
	case "reference":
		if media == 0 || frames != 0 {
			return nil, fmt.Errorf("reference requires images, audios, or videos and does not accept first_frame or last_frame")
		}
	}
	payload := make(map[string]any)
	for _, key := range []string{"prompt", "mode", "seconds", "size", "aspect_ratio", "seed", "n", "first_frame", "last_frame", "images", "audios", "videos"} {
		if value, exists := merged[key]; exists {
			payload[key] = value
		}
	}
	payload["model"] = modelName
	return payload, nil
}

func jsonInteger(value any) (int, bool) {
	switch value.(type) {
	case int, float64:
		return numberToInt(value)
	default:
		return 0, false
	}
}

func publicMediaURL(value any) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}

func validateURLArray(req map[string]any, key string, limit int) (int, error) {
	value, exists := req[key]
	if !exists || value == nil {
		return 0, nil
	}
	items, ok := value.([]any)
	if !ok {
		return 0, fmt.Errorf("%s must be an array of URLs", key)
	}
	if len(items) > limit {
		return 0, fmt.Errorf("%s length must not exceed %d", key, limit)
	}
	for _, item := range items {
		if !publicMediaURL(item) {
			return 0, fmt.Errorf("%s must contain HTTP URLs", key)
		}
	}
	return len(items), nil
}

func validateVideos(req map[string]any, flash bool) (int, error) {
	value, exists := req["videos"]
	if !exists || value == nil {
		return 0, nil
	}
	items, ok := value.([]any)
	if !ok {
		return 0, fmt.Errorf("videos must be an array")
	}
	if flash && len(items) != 0 {
		return 0, fmt.Errorf("videos is not supported")
	}
	if len(items) > 1 {
		return 0, fmt.Errorf("videos length must not exceed 1")
	}
	for _, item := range items {
		video, ok := item.(map[string]any)
		if !ok || !publicMediaURL(video["url"]) {
			return 0, fmt.Errorf("videos[].url must be an HTTP URL")
		}
		if value, exists := video["start_seconds"]; exists {
			seconds, ok := value.(float64)
			if !ok || seconds < 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
				return 0, fmt.Errorf("start_seconds must be a non-negative number")
			}
		}
		if value, exists := video["require_audio"]; exists {
			if _, ok := value.(bool); !ok {
				return 0, fmt.Errorf("require_audio must be a boolean")
			}
		}
	}
	return len(items), nil
}
