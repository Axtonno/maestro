package mutation

import (
	"errors"
	"strings"
)

var (
	ErrTargetNotFound  = errors.New("mutation target not found")
	ErrTargetAmbiguous = errors.New("mutation target is ambiguous")
	ErrStaleSource     = errors.New("mutation source is stale")
)

// CompileQualified adds M31 rejection classification without changing the
// frozen M28 compiler or the candidate it produces.
func CompileQualified(raw []byte, snapshot Snapshot) (Candidate, error) {
	proposal, err := Decode(raw)
	if err != nil {
		return Candidate{}, err
	}
	if proposal.Path != snapshot.Path {
		return Candidate{}, ErrPathMismatch
	}
	if !validText(snapshot.Content) || len(snapshot.Content) > MaxFileBytes {
		return Candidate{}, ErrPrecondition
	}
	if digest(snapshot.Content) != snapshot.Digest {
		return Candidate{}, errors.Join(ErrPrecondition, ErrStaleSource)
	}
	switch strings.Count(snapshot.Content, proposal.OldText) {
	case 0:
		return Candidate{}, errors.Join(ErrPrecondition, ErrTargetNotFound)
	case 1:
		return Compile(raw, snapshot)
	default:
		return Candidate{}, errors.Join(ErrPrecondition, ErrTargetAmbiguous)
	}
}

func TerminalForError(err error) string {
	switch {
	case errors.Is(err, ErrTargetNotFound):
		return "target_not_found"
	case errors.Is(err, ErrTargetAmbiguous):
		return "target_ambiguous"
	case errors.Is(err, ErrStaleSource):
		return "stale_source"
	case errors.Is(err, ErrSensitiveTarget):
		return "protected_target"
	default:
		return ""
	}
}
