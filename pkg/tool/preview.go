package tool

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	maxPreviewFields    = 16
	maxPreviewFieldSize = 4096
	maxPreviewBodyBytes = 256 << 10
)

// PreviewField is one redacted, human-readable metadata item attached to a
// prepared invocation. Labels are safe codes and values must fit on one line.
type PreviewField struct {
	label string
	value string
}

func NewPreviewField(label, value string) (PreviewField, error) {
	if !safeCode(label) || !exactValue(value, maxPreviewFieldSize) || strings.ContainsAny(value, "\r\n") {
		return PreviewField{}, fmt.Errorf("preview field is invalid: %w", ErrInvalidPreparedInvocation)
	}
	return PreviewField{label: label, value: value}, nil
}

func (field PreviewField) Label() string { return field.label }
func (field PreviewField) Value() string { return field.value }
func (field PreviewField) Validate() error {
	_, err := NewPreviewField(field.label, field.value)
	return err
}

// Preview is a bounded presentation prepared by a trusted tool before
// authorization. It is content-bound to the PreparedInvocation fingerprint.
type Preview struct {
	summary   string
	fields    []PreviewField
	body      string
	mediaType string
}

func NewPreview(summary string, fields []PreviewField, body, mediaType string) (Preview, error) {
	if !exactValue(summary, 512) || strings.ContainsAny(summary, "\r\n") ||
		len(fields) == 0 || len(fields) > maxPreviewFields ||
		body == "" || len(body) > maxPreviewBodyBytes || !utf8.ValidString(body) || strings.ContainsRune(body, 0) ||
		!exactValue(mediaType, 128) || strings.ContainsAny(mediaType, "\r\n \t") {
		return Preview{}, fmt.Errorf("prepared preview is invalid or exceeds its bounds: %w", ErrInvalidPreparedInvocation)
	}
	cloned := slices.Clone(fields)
	for _, field := range cloned {
		if err := field.Validate(); err != nil {
			return Preview{}, err
		}
	}
	return Preview{summary: summary, fields: cloned, body: body, mediaType: mediaType}, nil
}

func (preview Preview) Summary() string        { return preview.summary }
func (preview Preview) Fields() []PreviewField { return slices.Clone(preview.fields) }
func (preview Preview) Body() string           { return preview.body }
func (preview Preview) MediaType() string      { return preview.mediaType }

func (preview Preview) Validate() error {
	_, err := NewPreview(preview.summary, preview.fields, preview.body, preview.mediaType)
	return err
}

func clonePreview(preview Preview) Preview {
	preview.fields = slices.Clone(preview.fields)
	return preview
}
