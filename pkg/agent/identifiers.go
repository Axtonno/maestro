package agent

import (
	"cmp"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/antonio-cafeo/maestro/pkg/tool"
)

type ID string
type Version string
type StepID string
type RunID = tool.RunID

func (id ID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("agent ID %q is not namespaced: %w", id, ErrInvalidAgentID)
	}
	return nil
}

func (version Version) Validate() error {
	value := string(version)
	if !exactValue(value, 64) {
		return fmt.Errorf("agent version %q is not exact: %w", version, ErrInvalidVersion)
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("._+-", character) {
			continue
		}
		return fmt.Errorf("agent version %q contains unsupported characters: %w", version, ErrInvalidVersion)
	}
	return nil
}

func (id StepID) Validate() error {
	if !safeCode(string(id)) {
		return fmt.Errorf("plan step ID %q is not a safe code: %w", id, ErrInvalidStep)
	}
	return nil
}

func (id ID) Compare(other ID) int         { return cmp.Compare(id, other) }
func (id StepID) Compare(other StepID) int { return cmp.Compare(id, other) }

func namespacedID(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for index, character := range part {
			letter := character >= 'a' && character <= 'z'
			digit := character >= '0' && character <= '9'
			separator := index > 0 && (character == '_' || character == '-')
			if !letter && !digit && !separator {
				return false
			}
		}
	}
	return true
}

func exactValue(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value && !strings.ContainsRune(value, 0)
}

func safeCode(value string) bool {
	if !exactValue(value, 64) {
		return false
	}
	for index, character := range value {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		separator := index > 0 && (character == '_' || character == '-')
		if !letter && !digit && !separator {
			return false
		}
	}
	return true
}
