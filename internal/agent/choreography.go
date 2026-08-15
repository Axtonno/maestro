package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const (
	workspaceReadToolID  pkgTool.ID = "workspace.read"
	workspacePatchToolID pkgTool.ID = "workspace.patch"
)

type workspaceObservation struct {
	path    string
	digest  string
	content string
}

// toolChoreography keeps dependency-sensitive workspace mutations anchored to
// evidence produced by a completed tool turn. It does not grant authority and
// it never invokes a tool itself.
type toolChoreography struct {
	observation     *workspaceObservation
	pendingMutation pkgTool.ID
}

func (state *toolChoreography) toolsForTurn(
	tools []provider.Tool,
	descriptors map[string]pkgTool.Descriptor,
) []provider.Tool {
	available := make([]provider.Tool, 0, len(tools))
	for _, candidate := range tools {
		descriptor := descriptors[candidate.Name]
		if descriptor.ID() == workspacePatchToolID && state.observation == nil {
			continue
		}
		if state.pendingMutation != "" {
			if state.pendingMutation == workspacePatchToolID && state.observation == nil &&
				descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceMutate) {
				continue
			}
			if (state.pendingMutation != workspacePatchToolID || state.observation != nil) &&
				descriptor.ID() != state.pendingMutation {
				continue
			}
		}
		available = append(available, candidate)
	}
	return available
}

func (state *toolChoreography) beforeCall(
	descriptor pkgTool.Descriptor,
	arguments json.RawMessage,
	callIndex int,
) (pkgTool.Result, bool, error) {
	mutating := descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceMutate)
	if mutating {
		state.pendingMutation = descriptor.ID()
	}
	if callIndex > 0 {
		result, err := recoverableChoreographyResult("dependency_not_ready", "retry_next_turn")
		return result, false, err
	}
	if descriptor.ID() != workspacePatchToolID {
		return pkgTool.Result{}, true, nil
	}
	if state.observation == nil {
		result, err := recoverableChoreographyResult("dependency_not_ready", "workspace_read")
		return result, false, err
	}
	var patch struct {
		Path           string `json:"path"`
		Old            string `json:"old"`
		New            string `json:"new"`
		ExpectedDigest string `json:"expected_digest"`
	}
	if err := json.Unmarshal(arguments, &patch); err != nil {
		return pkgTool.Result{}, true, nil
	}
	if patch.Path != state.observation.path || patch.ExpectedDigest != state.observation.digest {
		result, err := recoverableChoreographyResult("stale_observation", "workspace_read")
		return result, false, err
	}
	if patch.Old == "" || !strings.Contains(state.observation.content, patch.Old) {
		result, err := recoverableChoreographyResult("invalid_observation", "select_exact_observed_text")
		return result, false, err
	}
	return pkgTool.Result{}, true, nil
}

func (state *toolChoreography) afterCall(descriptor pkgTool.Descriptor, result pkgTool.Result) {
	if descriptor.ID() == workspaceReadToolID {
		state.observation = verifiedWorkspaceObservation(result)
	}
	if !descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceMutate) {
		return
	}
	state.observation = nil
	if result.Outcome() == pkgTool.ResultSuccess {
		state.pendingMutation = ""
	}
}

func (state *toolChoreography) requiresMutation() bool { return state.pendingMutation != "" }

func verifiedWorkspaceObservation(result pkgTool.Result) *workspaceObservation {
	if result.Outcome() != pkgTool.ResultSuccess || result.MediaType() != "application/json" || result.Truncated() {
		return nil
	}
	var value struct {
		Path    string `json:"path"`
		Digest  string `json:"digest"`
		Content string `json:"content"`
	}
	if json.Unmarshal([]byte(result.Content()), &value) != nil || value.Path == "" || value.Digest == "" {
		return nil
	}
	digest, err := hex.DecodeString(value.Digest)
	if err != nil || len(digest) != sha256.Size {
		return nil
	}
	actual := sha256.Sum256([]byte(value.Content))
	if !strings.EqualFold(value.Digest, hex.EncodeToString(actual[:])) {
		return nil
	}
	return &workspaceObservation{path: value.Path, digest: value.Digest, content: value.Content}
}

func recoverableChoreographyResult(reason, requiredAction string) (pkgTool.Result, error) {
	content, err := json.Marshal(struct {
		Recoverable    bool   `json:"recoverable"`
		RequiredAction string `json:"required_action"`
	}{Recoverable: true, RequiredAction: requiredAction})
	if err != nil {
		return pkgTool.Result{}, err
	}
	return pkgTool.NewResult(pkgTool.ResultInvalid, string(content), "application/json", reason, 0, false, "")
}
