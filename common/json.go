package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func UnmarshalJsonStr(data string, v any) error {
	return json.Unmarshal(StringToByteSlice(data), v)
}

func DecodeJson(reader io.Reader, v any) error {
	return json.NewDecoder(reader).Decode(v)
}

// ValidateJSON validates exactly one JSON value without retaining the decoded
// value. It is intended for large response bodies that must be forwarded while
// still preserving the normal malformed-response error semantics.
func ValidateJSON(reader io.Reader) error {
	validator := jsonStreamValidator{}
	buffer := make([]byte, 32<<10)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			if validateErr := validator.feed(buffer[:n]); validateErr != nil {
				return validateErr
			}
		}
		if err == io.EOF {
			return validator.finish()
		}
		if err != nil {
			return err
		}
	}
}

type jsonContainer struct {
	kind  byte
	state uint8
}

const (
	jsonRootValue uint8 = iota
	jsonRootDone
	jsonObjectKeyOrEnd
	jsonObjectKey
	jsonObjectColon
	jsonObjectValue
	jsonObjectCommaOrEnd
	jsonArrayValueOrEnd
	jsonArrayValue
	jsonArrayCommaOrEnd
)

type jsonStreamValidator struct {
	stack       []jsonContainer
	rootState   uint8
	inString    bool
	stringKey   bool
	escape      bool
	unicodeLeft uint8
	literal     string
	literalPos  int
	numberState uint8
}

func (v *jsonStreamValidator) feed(data []byte) error {
	for _, b := range data {
		if err := v.feedByte(b); err != nil {
			return err
		}
	}
	return nil
}

func (v *jsonStreamValidator) feedByte(b byte) error {
	for {
		if v.inString {
			if v.unicodeLeft > 0 {
				if !isHexJSONByte(b) {
					return fmt.Errorf("invalid JSON unicode escape")
				}
				v.unicodeLeft--
				return nil
			}
			if v.escape {
				v.escape = false
				if b == 'u' {
					v.unicodeLeft = 4
					return nil
				}
				if !strings.ContainsRune(`"\\/bfnrt`, rune(b)) {
					return fmt.Errorf("invalid JSON escape")
				}
				return nil
			}
			switch b {
			case '\\':
				v.escape = true
			case '"':
				v.inString = false
				if v.stringKey {
					v.stringKey = false
					v.setState(jsonObjectColon)
				} else {
					v.valueComplete()
				}
			default:
				if b < 0x20 {
					return fmt.Errorf("invalid control character in JSON string")
				}
			}
			return nil
		}

		if v.literal != "" {
			if v.literalPos >= len(v.literal) {
				v.literal = ""
				v.valueComplete()
				continue
			}
			if b != v.literal[v.literalPos] {
				return fmt.Errorf("invalid JSON literal")
			}
			v.literalPos++
			return nil
		}

		if v.numberState != 0 {
			if v.feedNumberByte(b) {
				return nil
			}
			if v.numberState == 7 {
				v.numberState = 0
				v.valueComplete()
				continue
			}
			return fmt.Errorf("invalid JSON number")
		}

		if isJSONSpaceByte(b) {
			return nil
		}
		state := v.state()
		switch state {
		case jsonRootDone:
			return fmt.Errorf("multiple JSON values")
		case jsonObjectKeyOrEnd:
			if b == '}' {
				v.popContainer('}')
				return nil
			}
			if b != '"' {
				return fmt.Errorf("JSON object key must be a string")
			}
			v.inString, v.stringKey = true, true
			return nil
		case jsonObjectKey:
			if b != '"' {
				return fmt.Errorf("JSON object key must be a string")
			}
			v.inString, v.stringKey = true, true
			return nil
		case jsonObjectColon:
			if b != ':' {
				return fmt.Errorf("expected JSON object colon")
			}
			v.setState(jsonObjectValue)
			return nil
		case jsonObjectCommaOrEnd:
			if b == ',' {
				v.setState(jsonObjectKey)
				return nil
			}
			if b == '}' {
				v.popContainer('}')
				return nil
			}
			return fmt.Errorf("expected JSON object comma or end")
		case jsonArrayValueOrEnd:
			if b == ']' {
				v.popContainer(']')
				return nil
			}
			return v.startValue(b)
		case jsonArrayValue:
			return v.startValue(b)
		case jsonArrayCommaOrEnd:
			if b == ',' {
				v.setState(jsonArrayValue)
				return nil
			}
			if b == ']' {
				v.popContainer(']')
				return nil
			}
			return fmt.Errorf("expected JSON array comma or end")
		default:
			return v.startValue(b)
		}
	}
}

func (v *jsonStreamValidator) startValue(b byte) error {
	switch {
	case b == '{':
		v.stack = append(v.stack, jsonContainer{'{', jsonObjectKeyOrEnd})
	case b == '[':
		v.stack = append(v.stack, jsonContainer{'[', jsonArrayValueOrEnd})
	case b == '"':
		v.inString = true
	case b == 't':
		v.literal, v.literalPos = "true", 1
	case b == 'f':
		v.literal, v.literalPos = "false", 1
	case b == 'n':
		v.literal, v.literalPos = "null", 1
	case b == '-':
		v.numberState = 1
	case b >= '0' && b <= '9':
		if b == '0' {
			v.numberState = 2
		} else {
			v.numberState = 3
		}
	default:
		return fmt.Errorf("invalid JSON value")
	}
	return nil
}

func (v *jsonStreamValidator) feedNumberByte(b byte) bool {
	s := v.numberState
	if s == 1 {
		if b >= '1' && b <= '9' {
			v.numberState = 3
			return true
		}
		if b == '0' {
			v.numberState = 2
			return true
		}
		return false
	}
	if s == 2 {
		if b == '.' {
			v.numberState = 4
			return true
		}
		if b == 'e' || b == 'E' {
			v.numberState = 6
			return true
		}
		if isJSONDelimiter(b) {
			v.numberState = 7
			return false
		}
		return false
	}
	if s == 3 {
		if b >= '0' && b <= '9' {
			v.numberState = 3
			return true
		}
		if b == '.' {
			v.numberState = 4
			return true
		}
		if b == 'e' || b == 'E' {
			v.numberState = 6
			return true
		}
		if isJSONDelimiter(b) {
			v.numberState = 7
			return false
		}
		return false
	}
	if s == 4 {
		if b >= '0' && b <= '9' {
			v.numberState = 5
			return true
		}
		return false
	}
	if s == 5 {
		if b >= '0' && b <= '9' {
			return true
		}
		if b == 'e' || b == 'E' {
			v.numberState = 6
			return true
		}
		if isJSONDelimiter(b) {
			v.numberState = 7
			return false
		}
		return false
	}
	if s == 6 {
		if b == '+' || b == '-' {
			v.numberState = 8
			return true
		}
		if b >= '0' && b <= '9' {
			v.numberState = 9
			return true
		}
		return false
	}
	if s == 8 || s == 9 {
		if b >= '0' && b <= '9' {
			v.numberState = 9
			return true
		}
		if s == 9 && isJSONDelimiter(b) {
			v.numberState = 7
			return false
		}
	}
	return false
}

func (v *jsonStreamValidator) valueComplete() {
	if len(v.stack) == 0 {
		v.rootState = jsonRootDone
		return
	}
	container := &v.stack[len(v.stack)-1]
	if container.kind == '{' {
		container.state = jsonObjectCommaOrEnd
	} else {
		container.state = jsonArrayCommaOrEnd
	}
}

func (v *jsonStreamValidator) popContainer(expected byte) {
	wantKind := byte('{')
	if expected == ']' {
		wantKind = '['
	}
	if len(v.stack) == 0 || v.stack[len(v.stack)-1].kind != wantKind {
		return
	}
	v.stack = v.stack[:len(v.stack)-1]
	v.valueComplete()
}

func (v *jsonStreamValidator) state() uint8 {
	if len(v.stack) == 0 {
		return v.rootState
	}
	return v.stack[len(v.stack)-1].state
}

func (v *jsonStreamValidator) setState(state uint8) {
	if len(v.stack) == 0 {
		v.rootState = state
	} else {
		v.stack[len(v.stack)-1].state = state
	}
}

func (v *jsonStreamValidator) finish() error {
	if v.literal != "" && v.literalPos == len(v.literal) {
		v.literal = ""
		v.valueComplete()
	}
	if v.numberState == 2 || v.numberState == 3 || v.numberState == 5 || v.numberState == 9 {
		v.numberState = 0
		v.valueComplete()
	}
	if v.inString || v.escape || v.unicodeLeft > 0 || v.literal != "" || v.numberState != 0 || len(v.stack) != 0 || v.rootState != jsonRootDone {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func isJSONSpaceByte(b byte) bool { return b == ' ' || b == '\n' || b == '\r' || b == '\t' }
func isJSONDelimiter(b byte) bool { return isJSONSpaceByte(b) || b == ',' || b == ']' || b == '}' }
func isHexJSONByte(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func GetJsonType(data json.RawMessage) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "unknown"
	}
	firstChar := trimmed[0]
	switch firstChar {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		return "number"
	}
}

// JsonRawMessageToString returns JSON strings as their decoded value and other JSON values as raw text.
func JsonRawMessageToString(data json.RawMessage) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return ""
	}
	if trimmed[0] != '"' {
		return string(trimmed)
	}
	var value string
	if err := Unmarshal(trimmed, &value); err != nil {
		return string(trimmed)
	}
	return value
}
