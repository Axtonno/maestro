package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type Transport string

const (
	TransportNativeToolCall Transport = "native_tool_call"
	TransportStructured     Transport = "constrained_structured_output"
)

var ErrInvalidTransport = errors.New("invalid mutation transport output")

// DecodeTransport normalizes either frozen transport to the exact proposal
// bytes accepted by Decode and Compile. It never falls back to another
// transport and rejects prose or multiple JSON values.
func DecodeTransport(kind Transport, raw []byte) ([]byte, error) {
	switch kind {
	case TransportStructured:
		if _, err := Decode(raw); err != nil {
			return nil, errors.Join(ErrInvalidTransport, err)
		}
		return bytes.Clone(raw), nil
	case TransportNativeToolCall:
		if err := validateNativeCallKeys(raw); err != nil {
			return nil, err
		}
		var call struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&call); err != nil {
			return nil, errors.Join(ErrInvalidTransport, err)
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF || call.Name != "workspace_replace" {
			return nil, ErrInvalidTransport
		}
		if _, err := Decode(call.Arguments); err != nil {
			return nil, errors.Join(ErrInvalidTransport, err)
		}
		return bytes.Clone(call.Arguments), nil
	default:
		return nil, ErrInvalidTransport
	}
}

func validateNativeCallKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return ErrInvalidTransport
	}
	required := map[string]bool{"name": false, "arguments": false}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		seen, known := required[key]
		if err != nil || !ok || !known || seen {
			return ErrInvalidTransport
		}
		required[key] = true
		var discard json.RawMessage
		if err := decoder.Decode(&discard); err != nil {
			return ErrInvalidTransport
		}
	}
	if _, err := decoder.Token(); err != nil {
		return ErrInvalidTransport
	}
	for _, present := range required {
		if !present {
			return ErrInvalidTransport
		}
	}
	return nil
}
