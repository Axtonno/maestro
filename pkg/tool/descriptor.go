package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

type Descriptor struct {
	id          ID
	name        Name
	version     Version
	description string
	parameters  json.RawMessage
	effects     []Effect
}

func NewDescriptor(
	id ID,
	name Name,
	version Version,
	description string,
	parameters json.RawMessage,
	effects []Effect,
) (Descriptor, error) {
	if err := id.Validate(); err != nil {
		return Descriptor{}, fmt.Errorf("descriptor identity: %w: %w", err, ErrInvalidDescriptor)
	}
	if err := name.Validate(); err != nil {
		return Descriptor{}, fmt.Errorf("descriptor name: %w: %w", err, ErrInvalidDescriptor)
	}
	if err := version.Validate(); err != nil {
		return Descriptor{}, fmt.Errorf("descriptor version: %w: %w", err, ErrInvalidDescriptor)
	}
	if strings.TrimSpace(description) == "" || len(description) > 4096 ||
		!utf8.ValidString(description) || strings.ContainsRune(description, 0) {
		return Descriptor{}, fmt.Errorf("tool description is blank or unsafe: %w", ErrInvalidDescriptor)
	}
	normalized, err := normalizeJSONObject(parameters)
	if err != nil {
		return Descriptor{}, fmt.Errorf("tool parameters: %w: %w", err, ErrInvalidDescriptor)
	}
	clonedEffects := slices.Clone(effects)
	if len(clonedEffects) == 0 {
		return Descriptor{}, fmt.Errorf("tool must declare at least one possible effect: %w", ErrInvalidDescriptor)
	}
	seen := make(map[Effect]struct{}, len(clonedEffects))
	for _, effect := range clonedEffects {
		if !effect.ValidForTool() {
			return Descriptor{}, fmt.Errorf("tool effect %q is invalid: %w", effect, ErrInvalidDescriptor)
		}
		if _, exists := seen[effect]; exists {
			return Descriptor{}, fmt.Errorf("tool effect %q is duplicated: %w", effect, ErrInvalidDescriptor)
		}
		seen[effect] = struct{}{}
	}
	slices.Sort(clonedEffects)
	return Descriptor{
		id: id, name: name, version: version, description: description,
		parameters: normalized, effects: clonedEffects,
	}, nil
}

func (descriptor Descriptor) ID() ID                      { return descriptor.id }
func (descriptor Descriptor) Name() Name                  { return descriptor.name }
func (descriptor Descriptor) Version() Version            { return descriptor.version }
func (descriptor Descriptor) Description() string         { return descriptor.description }
func (descriptor Descriptor) Parameters() json.RawMessage { return bytes.Clone(descriptor.parameters) }
func (descriptor Descriptor) Effects() []Effect           { return slices.Clone(descriptor.effects) }

func (descriptor Descriptor) Validate() error {
	_, err := NewDescriptor(
		descriptor.id,
		descriptor.name,
		descriptor.version,
		descriptor.description,
		descriptor.parameters,
		descriptor.effects,
	)
	return err
}

func normalizeJSONObject(value json.RawMessage) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	var object map[string]any
	if len(bytes.TrimSpace(value)) == 0 || decoder.Decode(&object) != nil || object == nil {
		return nil, ErrInvalidInvocation
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, ErrInvalidInvocation
	}
	normalized, err := json.Marshal(object)
	if err != nil {
		return nil, ErrInvalidInvocation
	}
	return normalized, nil
}
