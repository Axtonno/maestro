package agent

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type RunResult struct {
	session SessionSnapshot
	content string
}

func NewRunResult(session SessionSnapshot, content string) (RunResult, error) {
	if err := session.Validate(); err != nil || session.State() != SessionTerminal {
		return RunResult{}, fmt.Errorf("run result requires a valid terminal session: %w", ErrInvalidResult)
	}
	if session.Terminal() == TerminalCompleted {
		if strings.TrimSpace(content) == "" || !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
			return RunResult{}, fmt.Errorf("completed run requires valid content: %w", ErrInvalidResult)
		}
	} else if !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
		return RunResult{}, fmt.Errorf("run result content is invalid: %w", ErrInvalidResult)
	}
	return RunResult{session: session, content: content}, nil
}

func (result RunResult) Session() SessionSnapshot { return result.session }
func (result RunResult) Content() string          { return result.content }

func (result RunResult) Validate() error {
	_, err := NewRunResult(result.session, result.content)
	return err
}
