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

// EffectState reports whether a mutating tool reached its commit point. The
// empty value means that the tool did not publish an effect state.
type EffectState string

const (
	EffectUnchanged EffectState = "unchanged"
	EffectApplied   EffectState = "applied"
)

func (state EffectState) Valid() bool {
	return state == EffectUnchanged || state == EffectApplied
}

type Result struct {
	outcome     ResultOutcome
	content     string
	mediaType   string
	reason      string
	itemCount   int
	truncated   bool
	disposition DenyDisposition
	effect      EffectState
	durable     bool
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
	return newResult(outcome, content, mediaType, reason, itemCount, truncated, disposition, "", false)
}

// NewEffectResult constructs a result that explicitly reports the mutation
// commit state. Durable is meaningful only after an applied effect.
func NewEffectResult(
	outcome ResultOutcome,
	content string,
	mediaType string,
	reason string,
	itemCount int,
	truncated bool,
	disposition DenyDisposition,
	effect EffectState,
	durable bool,
) (Result, error) {
	return newResult(outcome, content, mediaType, reason, itemCount, truncated, disposition, effect, durable)
}

func newResult(
	outcome ResultOutcome,
	content string,
	mediaType string,
	reason string,
	itemCount int,
	truncated bool,
	disposition DenyDisposition,
	effect EffectState,
	durable bool,
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
	if effect != "" && !effect.Valid() {
		return Result{}, fmt.Errorf("tool result effect state is invalid: %w", ErrInvalidResult)
	}
	if durable && effect != EffectApplied {
		return Result{}, fmt.Errorf("only an applied effect can be durable: %w", ErrInvalidResult)
	}
	if outcome == ResultDenied && effect != "" {
		return Result{}, fmt.Errorf("denied result cannot carry an effect state: %w", ErrInvalidResult)
	}
	return Result{
		outcome: outcome, content: content, mediaType: mediaType, reason: reason,
		itemCount: itemCount, truncated: truncated, disposition: disposition,
		effect: effect, durable: durable,
	}, nil
}

func (result Result) Outcome() ResultOutcome       { return result.outcome }
func (result Result) Content() string              { return result.content }
func (result Result) MediaType() string            { return result.mediaType }
func (result Result) Reason() string               { return result.reason }
func (result Result) ItemCount() int               { return result.itemCount }
func (result Result) Truncated() bool              { return result.truncated }
func (result Result) Disposition() DenyDisposition { return result.disposition }
func (result Result) Effect() EffectState          { return result.effect }
func (result Result) Durable() bool                { return result.durable }
func (result Result) Validate() error {
	_, err := newResult(
		result.outcome, result.content, result.mediaType, result.reason,
		result.itemCount, result.truncated, result.disposition, result.effect, result.durable,
	)
	return err
}
