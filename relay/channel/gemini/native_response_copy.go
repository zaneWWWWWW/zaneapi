package gemini

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// Small native generateContent bodies keep the original normalize+unmarshal
// path. Larger bodies are copied to the client without buffering the whole
// JSON tree; usageMetadata is extracted while copying.
const geminiNativeBufferLimit = 1 << 20

func serveGeminiNativeJSON(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)
	if resp == nil || resp.Body == nil {
		return nil, types.NewOpenAIError(io.ErrUnexpectedEOF, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	prefix, err := io.ReadAll(io.LimitReader(resp.Body, geminiNativeBufferLimit+1))
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if len(prefix) <= geminiNativeBufferLimit {
		return serveGeminiNativeBuffered(c, info, resp, prefix)
	}
	return serveGeminiNativeStreamed(c, info, resp, prefix)
}

func serveGeminiNativeBuffered(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response, responseBody []byte) (*dto.Usage, *types.NewAPIError) {
	logger.LogDebug(c, "Gemini native response body: %s", responseBody)
	responseBody, _, err := normalizeGeminiMarkdownImages(responseBody)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	var geminiResponse dto.GeminiChatResponse
	if err := common.Unmarshal(responseBody, &geminiResponse); err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if len(geminiResponse.Candidates) == 0 && geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
	}
	usage := buildUsageFromGeminiResponse(c, info, &geminiResponse)
	service.IOCopyBytesGracefully(c, resp, responseBody)
	return &usage, nil
}

func serveGeminiNativeStreamed(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response, prefix []byte) (*dto.Usage, *types.NewAPIError) {
	src := io.MultiReader(bytes.NewReader(prefix), resp.Body)
	probe := newGeminiBillingProbe()
	validationReader, validationWriter := io.Pipe()
	validationDone := make(chan error, 1)
	go func() {
		validationErr := common.ValidateJSON(validationReader)
		_, _ = io.Copy(io.Discard, validationReader)
		validationDone <- validationErr
	}()

	committed := false
	var client io.Writer
	if c.Writer != nil {
		copyGeminiUpstreamHeaders(c, resp)
		if resp.ContentLength > 0 {
			c.Writer.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		}
		status := resp.StatusCode
		if status == 0 {
			status = http.StatusOK
		}
		c.Writer.WriteHeader(status)
		committed = true
		client = c.Writer
	}

	clientErr, readErr := copyGeminiNativeBody(src, client, probe, validationWriter)
	if readErr != nil {
		_ = validationWriter.CloseWithError(readErr)
	} else {
		_ = validationWriter.Close()
	}
	validationErr := <-validationDone
	if clientErr != nil {
		logger.LogError(c, "failed to copy gemini native response body: "+clientErr.Error())
	}
	if validationErr != nil {
		logger.LogError(c, "gemini native response JSON validation failed: "+validationErr.Error())
	}
	if !committed {
		if readErr != nil {
			return nil, types.NewOpenAIError(readErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
		if validationErr != nil {
			return nil, types.NewOpenAIError(validationErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
	}
	if c.Writer != nil && clientErr == nil {
		c.Writer.Flush()
	}
	if probe.blockReason != "" {
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", probe.blockReason))
	}
	allowTextEstimate := readErr == nil && validationErr == nil
	usage := usageFromGeminiNativeProbe(c, info, probe, allowTextEstimate)
	return &usage, nil
}

// copyGeminiNativeBody forwards src to the client while always feeding the
// billing probe. A client write error does not stop the upstream drain, so
// usageMetadata at the end of a large generateContent body can still be billed
// after a broken pipe.
func copyGeminiNativeBody(src io.Reader, client io.Writer, probe *geminiBillingProbe, extra io.Writer) (clientErr error, readErr error) {
	if probe == nil {
		probe = newGeminiBillingProbe()
	}
	buf := make([]byte, 32<<10)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			probe.feed(chunk)
			if extra != nil {
				_, _ = extra.Write(chunk)
			}
			if client != nil && clientErr == nil {
				if _, writeErr := client.Write(chunk); writeErr != nil {
					clientErr = writeErr
				}
			}
		}
		if err == io.EOF {
			return clientErr, nil
		}
		if err != nil {
			return clientErr, err
		}
	}
}

func copyGeminiUpstreamHeaders(c *gin.Context, src *http.Response) {
	if c == nil || c.Writer == nil || src == nil {
		return
	}
	for k, v := range src.Header {
		if !service.ShouldCopyUpstreamHeader(c, k, v) {
			continue
		}
		c.Writer.Header().Set(k, v[0])
	}
}

func usageFromGeminiNativeProbe(c *gin.Context, info *relaycommon.RelayInfo, probe *geminiBillingProbe, allowTextEstimate bool) dto.Usage {
	if probe != nil && dto.HasGeminiUsageMetadataTokens(&probe.usage) {
		usage := buildUsageFromGeminiMetadata(&probe.usage, info.GetEstimatePromptTokens())
		patchGeminiZeroCompletionUsage(c, info, &usage, probe.text.String(), probe.imageCount)
		return usage
	}
	text := ""
	if allowTextEstimate && probe != nil {
		text = probe.text.String()
	}
	usage := service.ResponseText2Usage(c, text, info.UpstreamModelName, info.GetEstimatePromptTokens())
	attachEstimatedGeminiBillingUsage(usage)
	return *usage
}

type geminiBillingProbe struct {
	w io.Writer

	inString bool
	escape   bool
	depth    int

	readingKey bool
	keyBuf     []byte
	afterKey   bool
	currentKey string
	rootKey    string

	capturing    bool
	captureUntil int
	capture      []byte
	captureKind  string

	usage       dto.GeminiUsageMetadata
	blockReason string
	text        strings.Builder
	imageCount  int
}

func newGeminiBillingProbe() *geminiBillingProbe {
	return &geminiBillingProbe{}
}

func (p *geminiBillingProbe) Write(b []byte) (int, error) {
	if p.w == nil {
		p.w = io.Discard
	}
	n, err := p.w.Write(b)
	if n > 0 {
		p.feed(b[:n])
	}
	return n, err
}

func (p *geminiBillingProbe) feed(b []byte) {
	for i := 0; i < len(b); i++ {
		c := b[i]
		if p.inString {
			p.feedStringByte(c)
			continue
		}
		if p.capturing {
			p.capture = append(p.capture, c)
		}
		switch c {
		case '"':
			p.inString = true
			p.escape = false
			p.keyBuf = p.keyBuf[:0]
			p.readingKey = !p.afterKey
		case '{':
			p.depth++
			if p.afterKey && (p.currentKey == "inlineData" || p.currentKey == "inline_data") {
				p.imageCount++
			}
			if p.afterKey && !p.capturing {
				p.maybeStartCapture()
			}
			p.afterKey = false
		case '}':
			if p.capturing && p.depth == p.captureUntil {
				p.finishCapture()
			}
			p.depth--
			p.afterKey = false
		case '[':
			p.depth++
			p.afterKey = false
		case ']':
			p.depth--
			p.afterKey = false
		case ':':
			p.afterKey = true
		case ',':
			p.afterKey = false
		default:
			if p.afterKey && !isJSONSpace(c) && c != '{' && c != '[' && c != '"' {
				p.afterKey = false
			}
		}
	}
}

func (p *geminiBillingProbe) feedStringByte(c byte) {
	if p.capturing {
		p.capture = append(p.capture, c)
	}
	if p.escape {
		p.escape = false
		if p.readingKey {
			p.appendKeyByte(c)
		} else if p.currentKey == "blockReason" {
			p.appendKeyByte(c)
		} else if p.wantTextSample() {
			p.appendTextByte(c)
		}
		return
	}
	if c == '\\' {
		p.escape = true
		return
	}
	if c == '"' {
		p.inString = false
		p.finishString()
		return
	}
	if p.readingKey || p.currentKey == "blockReason" && p.afterKey {
		p.appendKeyByte(c)
	} else if p.wantTextSample() {
		p.appendTextByte(c)
	}
}

func (p *geminiBillingProbe) wantTextSample() bool {
	return p.afterKey && p.currentKey == "text" && p.rootKey != "usageMetadata" && p.text.Len() < 64<<10
}

func (p *geminiBillingProbe) appendKeyByte(c byte) {
	if len(p.keyBuf) < 64 {
		p.keyBuf = append(p.keyBuf, c)
	}
}

func (p *geminiBillingProbe) appendTextByte(c byte) {
	if p.text.Len() < 64<<10 {
		p.text.WriteByte(c)
	}
}

func (p *geminiBillingProbe) finishString() {
	if p.readingKey {
		p.currentKey = string(p.keyBuf)
		if p.depth == 1 {
			p.rootKey = p.currentKey
		}
		p.readingKey = false
		return
	}
	p.afterKey = false
	if p.currentKey == "blockReason" {
		p.blockReason = string(p.keyBuf)
	}
}

func (p *geminiBillingProbe) maybeStartCapture() {
	if p.depth != 2 {
		return
	}
	switch p.rootKey {
	case "usageMetadata":
		p.capturing = true
		p.captureKind = "usage"
		p.captureUntil = 2
		p.capture = append(p.capture[:0], '{')
	case "promptFeedback":
		p.capturing = true
		p.captureKind = "feedback"
		p.captureUntil = 2
		p.capture = append(p.capture[:0], '{')
	}
}

func (p *geminiBillingProbe) finishCapture() {
	p.capturing = false
	raw := p.capture
	p.capture = nil
	switch p.captureKind {
	case "usage":
		var metadata dto.GeminiUsageMetadata
		if err := common.Unmarshal(raw, &metadata); err == nil {
			p.usage = metadata
		}
	case "feedback":
		var feedback dto.GeminiChatPromptFeedback
		if err := common.Unmarshal(raw, &feedback); err == nil && feedback.BlockReason != nil {
			p.blockReason = *feedback.BlockReason
		}
	}
	p.captureKind = ""
}

func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}
