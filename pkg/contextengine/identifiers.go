package contextengine

import (
	"cmp"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

type WorkspaceID string
type SourceID string
type AnalyzerID string
type EstimatorID string
type DocumentPath string
type Digest string
type Language string

func (id WorkspaceID) Validate() error {
	if !exactID(string(id)) {
		return fmt.Errorf("workspace ID %q is not exact: %w", id, ErrInvalidWorkspaceID)
	}
	return nil
}

func (id SourceID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("source ID %q is not namespaced: %w", id, ErrInvalidSourceID)
	}
	return nil
}

func (id AnalyzerID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("analyzer ID %q is not namespaced: %w", id, ErrInvalidAnalyzerID)
	}
	return nil
}

func (id EstimatorID) Validate() error {
	if !namespacedID(string(id)) {
		return fmt.Errorf("estimator ID %q is not namespaced: %w", id, ErrInvalidEstimatorID)
	}
	return nil
}

func (documentPath DocumentPath) Validate() error {
	value := string(documentPath)
	if value == "" || value == "." || !utf8.ValidString(value) || strings.ContainsRune(value, 0) ||
		strings.Contains(value, "\\") || strings.HasPrefix(value, "/") || value == ".." ||
		strings.HasPrefix(value, "../") || path.Clean(value) != value {
		return fmt.Errorf("document path %q must be normalized and relative: %w", value, ErrInvalidPath)
	}
	return nil
}

func (digest Digest) Validate() error {
	value := string(digest)
	if len(value) != 64 {
		return fmt.Errorf("digest %q must contain 64 lowercase hexadecimal characters: %w", value, ErrInvalidDigest)
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return fmt.Errorf("digest %q must contain 64 lowercase hexadecimal characters: %w", value, ErrInvalidDigest)
		}
	}
	return nil
}

func (language Language) Validate() error {
	value := string(language)
	if value == "" {
		return nil
	}
	for index, character := range value {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		separator := index > 0 && (character == '+' || character == '#' || character == '-')
		if !letter && !digit && !separator {
			return fmt.Errorf("language %q is not normalized: %w", value, ErrInvalidDocument)
		}
	}
	return nil
}

func (id WorkspaceID) Compare(other WorkspaceID) int { return cmp.Compare(id, other) }
func (id SourceID) Compare(other SourceID) int       { return cmp.Compare(id, other) }
func (id AnalyzerID) Compare(other AnalyzerID) int   { return cmp.Compare(id, other) }
func (path DocumentPath) Compare(other DocumentPath) int {
	return cmp.Compare(path, other)
}

func exactID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && !strings.ContainsRune(value, 0)
}

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
