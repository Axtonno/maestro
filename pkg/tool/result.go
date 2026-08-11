package tool

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type ExecutionLimits struct {
	MaxDuration    time.Duration
	MaxOutputBytes int
	MaxItems       int
}

func (limits ExecutionLimits) Validate() error {
	if limits.MaxDuration <= 0 || limits.MaxDuration > 24*time.Hour ||
		limits.MaxOutputBytes <= 0 || limits.MaxOutputBytes > 64<<20 ||
		limits.MaxItems <= 0 || limits.MaxItems > 100_000 {
		return fmt.Errorf("tool limits must be positive and bounded: %w", ErrInvalidLimits)
	}
	return nil
}

type ResultOutcome string

const (
	ResultSuccess  ResultOutcome = "success"
	ResultDenied   ResultOutcome = "denied"
	ResultInvalid  ResultOutcome = "invalid"
	ResultFailed   ResultOutcome = "failed"
	ResultCanceled ResultOutcome = "canceled"
)

func (outcome ResultOutcome) Valid() bool {
	switch outcome {
	case ResultSuccess, ResultDenied, ResultInvalid, ResultFailed, ResultCanceled:
		return true
	default:
		return false
	}
}

type Result struct {
	outcome     ResultOutcome
	content     string
	mediaType   string
	reason      string
	itemCount   int
	truncated   bool
	disposition DenyDisposition
}

func NewResult(
	outcome ResultOutcome,
	content string,
	mediaType string,
	reason string,
	itemCount int,
	truncated bool,
	disposition DenyDisposition,
) (Result, error) {
	if !outcome.Valid() || !safeCode(reason) || itemCount < 0 || !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
		return Result{}, fmt.Errorf("tool result outcome, content, reason, or count is invalid: %w", ErrInvalidResult)
	}
	if content != "" && (!exactValue(mediaType, 128) || !strings.Contains(mediaType, "/")) {
		return Result{}, fmt.Errorf("tool result media type is invalid: %w", ErrInvalidResult)
	}
	if content == "" && mediaType != "" {
		return Result{}, fmt.Errorf("empty tool result cannot declare a media type: %w", ErrInvalidResult)
	}
	if outcome == ResultDenied {
		if disposition != DenyRecoverable && disposition != DenyTerminal {
			return Result{}, fmt.Errorf("denied result requires a disposition: %w", ErrInvalidResult)
		}
	} else if disposition != "" {
		return Result{}, fmt.Errorf("only denied results can carry a disposition: %w", ErrInvalidResult)
	}
	return Result{
		outcome: outcome, content: content, mediaType: mediaType, reason: reason,
		itemCount: itemCount, truncated: truncated, disposition: disposition,
	}, nil
}

func (result Result) Outcome() ResultOutcome       { return result.outcome }
func (result Result) Content() string              { return result.content }
func (result Result) MediaType() string            { return result.mediaType }
func (result Result) Reason() string               { return result.reason }
func (result Result) ItemCount() int               { return result.itemCount }
func (result Result) Truncated() bool              { return result.truncated }
func (result Result) Disposition() DenyDisposition { return result.disposition }
func (result Result) Validate() error {
	_, err := NewResult(
		result.outcome, result.content, result.mediaType, result.reason,
		result.itemCount, result.truncated, result.disposition,
	)
	return err
}
