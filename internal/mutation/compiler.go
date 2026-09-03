// Package mutation compiles untrusted Milestone 28 proposals into immutable,
// deterministic single-file replacement candidates. It deliberately owns no
// filesystem or provider authority.
package mutation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"
)

const (
	Version          = 1
	OperationReplace = "replace"
	MaxProposalBytes = 256 << 10
	MaxFileBytes     = 2 << 20
)

var (
	ErrInvalidProposal = errors.New("invalid mutation proposal")
	ErrPathMismatch    = errors.New("proposal path does not match authoritative snapshot")
	ErrSensitiveTarget = errors.New("sensitive or excluded mutation target")
	ErrPrecondition    = errors.New("mutation precondition failed")
)

// Snapshot is the complete authoritative read made in the same run. Digest is
// checked rather than trusted, so callers cannot bind a proposal to truncated
// or incorrectly labelled content.
type Snapshot struct {
	Path    string
	Content string
	Digest  string
}

// ProposalV1 is the only provider-facing M28 representation.
type ProposalV1 struct {
	Version   int    `json:"version"`
	Path      string `json:"path"`
	Operation string `json:"operation"`
	OldText   string `json:"old_text"`
	NewText   string `json:"new_text"`
}

// Candidate is the compiled unit used for preview, approval and application.
// Its fields are private so later stages cannot alter the approved action.
type Candidate struct {
	path, before, after, oldText, newText  string
	beforeDigest, afterDigest, fingerprint string
}

func (c Candidate) Path() string         { return c.path }
func (c Candidate) Before() string       { return c.before }
func (c Candidate) After() string        { return c.after }
func (c Candidate) OldText() string      { return c.oldText }
func (c Candidate) NewText() string      { return c.newText }
func (c Candidate) BeforeDigest() string { return c.beforeDigest }
func (c Candidate) AfterDigest() string  { return c.afterDigest }
func (c Candidate) Fingerprint() string  { return c.fingerprint }

// Decode rejects unknown, missing and duplicate fields without repair. It can
// be used before an authoritative read so excluded targets are never opened.
func Decode(raw []byte) (ProposalV1, error) {
	if len(raw) == 0 || len(raw) > MaxProposalBytes || !utf8.Valid(raw) {
		return ProposalV1{}, ErrInvalidProposal
	}
	if err := validateObjectKeys(raw); err != nil {
		return ProposalV1{}, err
	}
	var proposal ProposalV1
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proposal); err != nil {
		return ProposalV1{}, fmt.Errorf("decode proposal: %w: %w", err, ErrInvalidProposal)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ProposalV1{}, ErrInvalidProposal
	}
	if proposal.Version != Version || proposal.Operation != OperationReplace ||
		proposal.Path == "" || proposal.OldText == "" || proposal.OldText == proposal.NewText ||
		!validText(proposal.OldText) || !validText(proposal.NewText) {
		return ProposalV1{}, ErrInvalidProposal
	}
	if err := validateLogicalPath(proposal.Path); err != nil {
		return ProposalV1{}, err
	}
	return proposal, nil
}

// Compile binds a decoded proposal to the authoritative snapshot.
func Compile(raw []byte, snapshot Snapshot) (Candidate, error) {
	proposal, err := Decode(raw)
	if err != nil {
		return Candidate{}, err
	}
	if proposal.Path != snapshot.Path {
		return Candidate{}, ErrPathMismatch
	}
	if !validText(snapshot.Content) || len(snapshot.Content) > MaxFileBytes || digest(snapshot.Content) != snapshot.Digest {
		return Candidate{}, ErrPrecondition
	}
	if strings.Count(snapshot.Content, proposal.OldText) != 1 {
		return Candidate{}, ErrPrecondition
	}
	after := strings.Replace(snapshot.Content, proposal.OldText, proposal.NewText, 1)
	if len(after) > MaxFileBytes {
		return Candidate{}, ErrInvalidProposal
	}
	beforeDigest, afterDigest := digest(snapshot.Content), digest(after)
	fingerprintInput, _ := json.Marshal(struct {
		Version, Operation, Path, BeforeDigest, AfterDigest, OldText, NewText any
	}{Version, OperationReplace, proposal.Path, beforeDigest, afterDigest, proposal.OldText, proposal.NewText})
	return Candidate{
		path: proposal.Path, before: snapshot.Content, after: after,
		oldText: proposal.OldText, newText: proposal.NewText,
		beforeDigest: beforeDigest, afterDigest: afterDigest,
		fingerprint: digest(string(fingerprintInput)),
	}, nil
}

func validateObjectKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return ErrInvalidProposal
	}
	required := map[string]bool{"version": false, "path": false, "operation": false, "old_text": false, "new_text": false}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return ErrInvalidProposal
		}
		seen, known := required[key]
		if !known || seen {
			return ErrInvalidProposal
		}
		required[key] = true
		var discard json.RawMessage
		if err := decoder.Decode(&discard); err != nil {
			return ErrInvalidProposal
		}
	}
	if _, err := decoder.Token(); err != nil {
		return ErrInvalidProposal
	}
	for _, present := range required {
		if !present {
			return ErrInvalidProposal
		}
	}
	return nil
}

func validateLogicalPath(value string) error {
	if value == "" || value != path.Clean(value) || path.IsAbs(value) || strings.Contains(value, "\\") || strings.ContainsRune(value, 0) {
		return ErrInvalidProposal
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "." || part == ".." {
			return ErrInvalidProposal
		}
		lower := strings.ToLower(part)
		if strings.HasPrefix(part, ".") ||
			lower == "credential" || lower == "credentials" || lower == "secret" || lower == "secrets" ||
			lower == "vendor" || lower == "dist" || lower == "build" || lower == "generated" ||
			strings.HasPrefix(lower, ".env") || strings.Contains(lower, ".generated.") {
			return ErrSensitiveTarget
		}
	}
	return nil
}

func validText(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
