package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type BinaryDecision string

const (
	BinaryPropose BinaryDecision = "propose"
	BinaryAbstain BinaryDecision = "abstain"
)

var ErrInvalidBinaryDecision = errors.New("invalid binary mutation decision")

type BinaryDecisionV1 struct {
	Decision BinaryDecision
	Proposal []byte
}

// DecodeBinaryDecision accepts only the two M32 authoritative outcomes. A
// proposal is normalized to the frozen mutation-proposal-v1 representation.
func DecodeBinaryDecision(raw []byte) (BinaryDecisionV1, error) {
	if len(raw) == 0 || len(raw) > MaxProposalBytes {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return BinaryDecisionV1{}, ErrInvalidBinaryDecision
		}
		if _, duplicate := fields[key]; duplicate {
			return BinaryDecisionV1{}, ErrInvalidBinaryDecision
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return BinaryDecisionV1{}, ErrInvalidBinaryDecision
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	var version int
	var decision BinaryDecision
	if json.Unmarshal(fields["version"], &version) != nil || version != Version || json.Unmarshal(fields["decision"], &decision) != nil {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	if decision == BinaryAbstain {
		if len(fields) != 2 {
			return BinaryDecisionV1{}, ErrInvalidBinaryDecision
		}
		return BinaryDecisionV1{Decision: decision}, nil
	}
	if decision != BinaryPropose || len(fields) != 6 {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	proposalFields := map[string]json.RawMessage{}
	for _, key := range []string{"version", "path", "operation", "old_text", "new_text"} {
		value, exists := fields[key]
		if !exists {
			return BinaryDecisionV1{}, ErrInvalidBinaryDecision
		}
		proposalFields[key] = value
	}
	proposal, err := json.Marshal(proposalFields)
	if err != nil {
		return BinaryDecisionV1{}, ErrInvalidBinaryDecision
	}
	if _, err := Decode(proposal); err != nil {
		return BinaryDecisionV1{}, errors.Join(ErrInvalidBinaryDecision, err)
	}
	return BinaryDecisionV1{Decision: decision, Proposal: proposal}, nil
}
