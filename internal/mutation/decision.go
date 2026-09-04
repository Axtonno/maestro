package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type Decision string

const (
	DecisionPropose                   Decision = "propose"
	DecisionAbstainMissingInformation Decision = "abstain_missing_information"
	DecisionAbstainTargetNotFound     Decision = "abstain_target_not_found"
	DecisionAbstainTargetAmbiguous    Decision = "abstain_target_ambiguous"
)

var ErrInvalidDecision = errors.New("invalid mutation decision")

type StructuredDecision struct {
	Version  int
	Decision Decision
	Proposal []byte
}

// DecodeDecision validates the strict M30 envelope. It does not repair output
// and delegates proposal validation to the frozen mutation-proposal-v1 decoder.
func DecodeDecision(raw []byte) (StructuredDecision, error) {
	if len(raw) == 0 || len(raw) > MaxProposalBytes {
		return StructuredDecision{}, ErrInvalidDecision
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return StructuredDecision{}, ErrInvalidDecision
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || (key != "version" && key != "decision" && key != "proposal") {
			return StructuredDecision{}, ErrInvalidDecision
		}
		if _, duplicate := fields[key]; duplicate {
			return StructuredDecision{}, ErrInvalidDecision
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return StructuredDecision{}, ErrInvalidDecision
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return StructuredDecision{}, ErrInvalidDecision
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return StructuredDecision{}, ErrInvalidDecision
	}
	var version int
	var decision Decision
	if json.Unmarshal(fields["version"], &version) != nil || version != Version ||
		json.Unmarshal(fields["decision"], &decision) != nil {
		return StructuredDecision{}, ErrInvalidDecision
	}
	result := StructuredDecision{Version: version, Decision: decision}
	if decision == DecisionPropose {
		proposal, exists := fields["proposal"]
		if !exists || len(fields) != 3 {
			return StructuredDecision{}, ErrInvalidDecision
		}
		if _, err := Decode(proposal); err != nil {
			return StructuredDecision{}, errors.Join(ErrInvalidDecision, err)
		}
		result.Proposal = bytes.Clone(proposal)
		return result, nil
	}
	if len(fields) != 2 {
		return StructuredDecision{}, ErrInvalidDecision
	}
	switch decision {
	case DecisionAbstainMissingInformation, DecisionAbstainTargetNotFound, DecisionAbstainTargetAmbiguous:
		return result, nil
	default:
		return StructuredDecision{}, ErrInvalidDecision
	}
}
