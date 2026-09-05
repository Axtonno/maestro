package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

var ErrSelectionOutOfBounds = errors.New("selection_out_of_bounds")
var ErrInsufficientInformation = errors.New("insufficient_information")

// ValidateTarget applies the same excluded-path rules before a host read.
func ValidateTarget(path string) error { return validateLogicalPath(path) }

// Selection freezes host-owned coordinates and bytes before generation.
type Selection struct {
	snapshot             Snapshot
	start, end, from, to int
}

func Select(snapshot Snapshot, start, end int) (Selection, error) {
	if err := validateLogicalPath(snapshot.Path); err != nil {
		return Selection{}, err
	}
	if !validText(snapshot.Content) || len(snapshot.Content) > MaxFileBytes || digest(snapshot.Content) != snapshot.Digest {
		return Selection{}, ErrPrecondition
	}
	if start < 1 || end < start || snapshot.Content == "" {
		return Selection{}, ErrSelectionOutOfBounds
	}
	lines := strings.SplitAfter(snapshot.Content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if end > len(lines) {
		return Selection{}, ErrSelectionOutOfBounds
	}
	from := 0
	for _, line := range lines[:start-1] {
		from += len(line)
	}
	to := from
	for _, line := range lines[start-1 : end] {
		to += len(line)
	}
	if snapshot.Content[to-1] == '\n' {
		to--
		if to > from && snapshot.Content[to-1] == '\r' {
			to--
		}
	}
	return Selection{snapshot, start, end, from, to}, nil
}

func (s Selection) Path() string         { return s.snapshot.Path }
func (s Selection) Before() string       { return s.snapshot.Content }
func (s Selection) BeforeDigest() string { return s.snapshot.Digest }
func (s Selection) Text() string         { return s.snapshot.Content[s.from:s.to] }
func (s Selection) StartLine() int       { return s.start }
func (s Selection) EndLine() int         { return s.end }

func (s Selection) Replace(text string) (string, error) {
	if s.start < 1 || !validText(text) || text == s.Text() {
		return "", ErrInvalidProposal
	}
	after := s.Before()[:s.from] + text + s.Before()[s.to:]
	if len(after) > MaxFileBytes {
		return "", ErrInvalidProposal
	}
	return after, nil
}

// Fingerprint binds the exact rendered preview as well as both byte spans.
func (s Selection) Fingerprint(replacement, diff string) string {
	encoded, _ := json.Marshal([]any{"host-bound-mutation-decision-v1", s.Path(), s.BeforeDigest(), s.start, s.end, digest(s.Text()), digest(replacement), digest(diff)})
	return digest(string(encoded))
}

type HostBoundDecision struct {
	Decision BinaryDecision
	NewText  string
}

// DecodeHostBoundDecision rejects duplicate, unknown, null and trailing fields.
func DecodeHostBoundDecision(raw []byte) (HostBoundDecision, error) {
	invalid := func() (HostBoundDecision, error) { return HostBoundDecision{}, ErrInvalidBinaryDecision }
	if len(raw) == 0 || len(raw) > MaxProposalBytes || !validText(string(raw)) {
		return invalid()
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return invalid()
	}
	fields := map[string]string{}
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		if err != nil || !ok || (key != "decision" && key != "new_text") {
			return invalid()
		}
		if _, exists := fields[key]; exists {
			return invalid()
		}
		var value *string
		if d.Decode(&value) != nil || value == nil || !validText(*value) {
			return invalid()
		}
		fields[key] = *value
	}
	if _, err := d.Token(); err != nil {
		return invalid()
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return invalid()
	}
	if fields["decision"] == "abstain" && len(fields) == 1 {
		return HostBoundDecision{Decision: BinaryAbstain}, nil
	}
	if fields["decision"] == "propose" && len(fields) == 2 {
		return HostBoundDecision{BinaryPropose, fields["new_text"]}, nil
	}
	return invalid()
}
