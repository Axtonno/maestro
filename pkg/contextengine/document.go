package contextengine

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode/utf8"
)

type SourceRange struct {
	Start int
	End   int
}

func (sourceRange SourceRange) Validate(size int) error {
	if size < 0 || sourceRange.Start < 0 || sourceRange.End <= sourceRange.Start || sourceRange.End > size {
		return fmt.Errorf("source range [%d,%d) exceeds content size %d: %w", sourceRange.Start, sourceRange.End, size, ErrInvalidRange)
	}
	return nil
}

// Document is immutable content identified by a normalized logical path and a
// SHA-256 digest. Text documents must contain normalized UTF-8; opaque content
// uses application/octet-stream and is indexable but not analyzable as text.
type Document struct {
	path      DocumentPath
	digest    Digest
	mediaType string
	language  Language
	content   string
}

func NewDocument(documentPath DocumentPath, mediaType string, language Language, content string) (Document, error) {
	if err := documentPath.Validate(); err != nil {
		return Document{}, fmt.Errorf("document identity: %w: %w", err, ErrInvalidDocument)
	}
	if mediaType == "" || strings.TrimSpace(mediaType) != mediaType || !strings.Contains(mediaType, "/") {
		return Document{}, fmt.Errorf("document media type %q is invalid: %w", mediaType, ErrInvalidDocument)
	}
	if err := language.Validate(); err != nil {
		return Document{}, err
	}
	textual := strings.HasPrefix(mediaType, "text/") || language != ""
	if textual && (!utf8.ValidString(content) || strings.ContainsRune(content, 0)) {
		return Document{}, fmt.Errorf("document content is not supported UTF-8 text: %w", ErrInvalidDocument)
	}
	if !textual && mediaType != "application/octet-stream" {
		return Document{}, fmt.Errorf("opaque document media type %q is unsupported: %w", mediaType, ErrInvalidDocument)
	}
	sum := sha256.Sum256([]byte(content))
	return Document{
		path: documentPath, digest: Digest(fmt.Sprintf("%x", sum)),
		mediaType: mediaType, language: language, content: content,
	}, nil
}

func (document Document) Path() DocumentPath { return document.path }
func (document Document) Digest() Digest     { return document.digest }
func (document Document) MediaType() string  { return document.mediaType }
func (document Document) Language() Language { return document.language }
func (document Document) Content() string    { return document.content }
func (document Document) SizeBytes() int     { return len(document.content) }

func (document Document) Validate() error {
	constructed, err := NewDocument(document.path, document.mediaType, document.language, document.content)
	if err != nil {
		return err
	}
	if document.digest != constructed.digest {
		return fmt.Errorf("document digest does not match content: %w: %w", ErrInvalidDigest, ErrInvalidDocument)
	}
	return nil
}

type DiagnosticSeverity string

const (
	DiagnosticInfo    DiagnosticSeverity = "info"
	DiagnosticWarning DiagnosticSeverity = "warning"
	DiagnosticError   DiagnosticSeverity = "error"
)

func (severity DiagnosticSeverity) Valid() bool {
	return severity == DiagnosticInfo || severity == DiagnosticWarning || severity == DiagnosticError
}

type Diagnostic struct {
	Path     DocumentPath
	Range    SourceRange
	Severity DiagnosticSeverity
	Code     string
}

func (diagnostic Diagnostic) Validate(size int) error {
	if err := diagnostic.Path.Validate(); err != nil {
		return fmt.Errorf("diagnostic path: %w: %w", err, ErrInvalidDiagnostic)
	}
	if diagnostic.Range != (SourceRange{}) {
		if err := diagnostic.Range.Validate(size); err != nil {
			return fmt.Errorf("diagnostic range: %w: %w", err, ErrInvalidDiagnostic)
		}
	}
	if !diagnostic.Severity.Valid() || !safeCode(diagnostic.Code) {
		return fmt.Errorf("diagnostic severity %q or code %q is invalid: %w", diagnostic.Severity, diagnostic.Code, ErrInvalidDiagnostic)
	}
	return nil
}

func safeCode(value string) bool {
	if value == "" || len(value) > 64 {
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
