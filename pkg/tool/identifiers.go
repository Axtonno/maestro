package tool

import (
	"cmp"
	"fmt"
	"strings"
	"unicode/utf8"
)

type ID string
type Name string
type PolicyID string
type RunID string
type CallID string
type Version string
type Fingerprint string

func (id ID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("tool ID %q is not namespaced: %w", id, ErrInvalidToolID)
	}
	return nil
}

func (name Name) Validate() error {
	value := string(name)
	if len(value) == 0 || len(value) > 64 || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return fmt.Errorf("tool name %q is not provider-compatible: %w", name, ErrInvalidToolName)
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '-' {
			continue
		}
		return fmt.Errorf("tool name %q is not provider-compatible: %w", name, ErrInvalidToolName)
	}
	return nil
}

func (id PolicyID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("policy ID %q is not namespaced: %w", id, ErrInvalidPolicyID)
	}
	return nil
}

func (id RunID) Validate() error {
	if !exactValue(string(id), 128) || strings.ContainsAny(string(id), "\r\n") {
		return fmt.Errorf("run ID %q is not exact: %w", id, ErrInvalidRunID)
	}
	return nil
}

func (id CallID) Validate() error {
	if !exactValue(string(id), 256) || strings.ContainsAny(string(id), "\r\n") {
		return fmt.Errorf("call ID %q is not exact: %w", id, ErrInvalidCallID)
	}
	return nil
}

func (version Version) Validate() error {
	value := string(version)
	if !exactValue(value, 64) {
		return fmt.Errorf("tool version %q is not exact: %w", version, ErrInvalidVersion)
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("._+-", character) {
			continue
		}
		return fmt.Errorf("tool version %q contains unsupported characters: %w", version, ErrInvalidVersion)
	}
	return nil
}

func (fingerprint Fingerprint) Validate() error {
	if !lowerHex(string(fingerprint), 64) {
		return fmt.Errorf("fingerprint %q is not a SHA-256 digest: %w", fingerprint, ErrInvalidPreparedInvocation)
	}
	return nil
}

func (id ID) Compare(other ID) int             { return cmp.Compare(id, other) }
func (name Name) Compare(other Name) int       { return cmp.Compare(name, other) }
func (id PolicyID) Compare(other PolicyID) int { return cmp.Compare(id, other) }

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

func lowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
