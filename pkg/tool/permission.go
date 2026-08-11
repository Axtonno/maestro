package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/provider"
)

type PermissionSubject string

const (
	PermissionSubjectTool  PermissionSubject = "tool"
	PermissionSubjectModel PermissionSubject = "model"
)

func (subject PermissionSubject) Valid() bool {
	return subject == PermissionSubjectTool || subject == PermissionSubjectModel
}

type ModelTarget struct {
	provider provider.ID
	model    string
}

func NewModelTarget(providerID provider.ID, model string) (ModelTarget, error) {
	if !exactValue(string(providerID), 256) || strings.ContainsAny(string(providerID), "\r\n") ||
		!exactValue(model, 512) || strings.ContainsAny(model, "\r\n") {
		return ModelTarget{}, fmt.Errorf("provider and model must be exact: %w", ErrInvalidPermissionRequest)
	}
	return ModelTarget{provider: providerID, model: model}, nil
}

func (target ModelTarget) Provider() provider.ID { return target.provider }
func (target ModelTarget) Model() string         { return target.model }
func (target ModelTarget) Validate() error {
	_, err := NewModelTarget(target.provider, target.model)
	return err
}

type DisclosureManifest struct {
	workspace   contextengine.WorkspaceID
	generation  uint64
	sections    int
	tokens      int
	bytes       int64
	fingerprint Fingerprint
}

func NewDisclosureManifest(
	workspace contextengine.WorkspaceID,
	generation uint64,
	sections int,
	tokens int,
	bytes int64,
	fingerprint Fingerprint,
) (DisclosureManifest, error) {
	if err := workspace.Validate(); err != nil || generation == 0 || sections <= 0 ||
		tokens <= 0 || bytes <= 0 {
		return DisclosureManifest{}, fmt.Errorf("disclosure metadata is invalid: %w", ErrInvalidPermissionRequest)
	}
	if err := fingerprint.Validate(); err != nil {
		return DisclosureManifest{}, fmt.Errorf("disclosure fingerprint: %w: %w", err, ErrInvalidPermissionRequest)
	}
	return DisclosureManifest{
		workspace: workspace, generation: generation, sections: sections,
		tokens: tokens, bytes: bytes, fingerprint: fingerprint,
	}, nil
}

func (manifest DisclosureManifest) Workspace() contextengine.WorkspaceID { return manifest.workspace }
func (manifest DisclosureManifest) Generation() uint64                   { return manifest.generation }
func (manifest DisclosureManifest) Sections() int                        { return manifest.sections }
func (manifest DisclosureManifest) Tokens() int                          { return manifest.tokens }
func (manifest DisclosureManifest) Bytes() int64                         { return manifest.bytes }
func (manifest DisclosureManifest) Fingerprint() Fingerprint             { return manifest.fingerprint }

func (manifest DisclosureManifest) Validate() error {
	_, err := NewDisclosureManifest(
		manifest.workspace,
		manifest.generation,
		manifest.sections,
		manifest.tokens,
		manifest.bytes,
		manifest.fingerprint,
	)
	return err
}

type PermissionRequest struct {
	policy      PolicyID
	run         RunID
	subject     PermissionSubject
	actions     []Action
	fingerprint Fingerprint
	prepared    *PreparedInvocation
	target      *ModelTarget
	disclosure  *DisclosureManifest
}

func NewToolPermissionRequest(policy PolicyID, prepared PreparedInvocation) (PermissionRequest, error) {
	if err := policy.Validate(); err != nil {
		return PermissionRequest{}, fmt.Errorf("permission policy: %w: %w", err, ErrInvalidPermissionRequest)
	}
	if err := prepared.Validate(); err != nil {
		return PermissionRequest{}, fmt.Errorf("permission invocation: %w: %w", err, ErrInvalidPermissionRequest)
	}
	copyPrepared := prepared
	fingerprint := fingerprintToolPermission(policy, prepared)
	return PermissionRequest{
		policy: policy, run: prepared.invocation.run, subject: PermissionSubjectTool,
		actions: prepared.Actions(), fingerprint: fingerprint, prepared: &copyPrepared,
	}, nil
}

func NewModelPermissionRequest(
	policy PolicyID,
	run RunID,
	target ModelTarget,
	disclosure *DisclosureManifest,
) (PermissionRequest, error) {
	if err := policy.Validate(); err != nil {
		return PermissionRequest{}, fmt.Errorf("permission policy: %w: %w", err, ErrInvalidPermissionRequest)
	}
	if err := run.Validate(); err != nil {
		return PermissionRequest{}, fmt.Errorf("permission run: %w: %w", err, ErrInvalidPermissionRequest)
	}
	if err := target.Validate(); err != nil {
		return PermissionRequest{}, err
	}
	actions := make([]Action, 0, 2)
	invoke, _ := NewAction(EffectModelInvoke, string(target.provider)+"/"+target.model, "")
	actions = append(actions, invoke)
	var copiedDisclosure *DisclosureManifest
	if disclosure != nil {
		if err := disclosure.Validate(); err != nil {
			return PermissionRequest{}, fmt.Errorf("permission disclosure: %w: %w", err, ErrInvalidPermissionRequest)
		}
		disclose, _ := NewAction(
			EffectModelDisclose,
			string(disclosure.fingerprint),
			disclosure.workspace,
		)
		actions = append(actions, disclose)
		copyManifest := *disclosure
		copiedDisclosure = &copyManifest
	}
	fingerprint := fingerprintPermission(policy, run, PermissionSubjectModel, actions)
	copyTarget := target
	return PermissionRequest{
		policy: policy, run: run, subject: PermissionSubjectModel, actions: actions,
		fingerprint: fingerprint, target: &copyTarget, disclosure: copiedDisclosure,
	}, nil
}

func (request PermissionRequest) Policy() PolicyID           { return request.policy }
func (request PermissionRequest) Run() RunID                 { return request.run }
func (request PermissionRequest) Subject() PermissionSubject { return request.subject }
func (request PermissionRequest) Actions() []Action          { return slices.Clone(request.actions) }
func (request PermissionRequest) Fingerprint() Fingerprint   { return request.fingerprint }
func (request PermissionRequest) Prepared() (PreparedInvocation, bool) {
	if request.prepared == nil {
		return PreparedInvocation{}, false
	}
	return *request.prepared, true
}
func (request PermissionRequest) ModelTarget() (ModelTarget, bool) {
	if request.target == nil {
		return ModelTarget{}, false
	}
	return *request.target, true
}
func (request PermissionRequest) Disclosure() (DisclosureManifest, bool) {
	if request.disclosure == nil {
		return DisclosureManifest{}, false
	}
	return *request.disclosure, true
}

func (request PermissionRequest) Validate() error {
	switch request.subject {
	case PermissionSubjectTool:
		if request.prepared == nil || request.target != nil || request.disclosure != nil {
			return ErrInvalidPermissionRequest
		}
		rebuilt, err := NewToolPermissionRequest(request.policy, *request.prepared)
		if err != nil || rebuilt.fingerprint != request.fingerprint {
			return fmt.Errorf("tool permission request is inconsistent: %w", ErrInvalidPermissionRequest)
		}
	case PermissionSubjectModel:
		if request.prepared != nil || request.target == nil {
			return ErrInvalidPermissionRequest
		}
		rebuilt, err := NewModelPermissionRequest(request.policy, request.run, *request.target, request.disclosure)
		if err != nil || rebuilt.fingerprint != request.fingerprint {
			return fmt.Errorf("model permission request is inconsistent: %w", ErrInvalidPermissionRequest)
		}
	default:
		return ErrInvalidPermissionRequest
	}
	return nil
}

func fingerprintPermission(policy PolicyID, run RunID, subject PermissionSubject, actions []Action) Fingerprint {
	hash := sha256.New()
	writeFingerprintPart(hash, string(policy))
	writeFingerprintPart(hash, string(run))
	writeFingerprintPart(hash, string(subject))
	for _, action := range actions {
		writeFingerprintPart(hash, string(action.effect))
		writeFingerprintPart(hash, action.resource)
		writeFingerprintPart(hash, string(action.workspace))
	}
	return Fingerprint(hex.EncodeToString(hash.Sum(nil)))
}

type DecisionKind string

const (
	DecisionAllow  DecisionKind = "allow"
	DecisionDeny   DecisionKind = "deny"
	DecisionPrompt DecisionKind = "prompt"
)

type DenyDisposition string

const (
	DenyRecoverable DenyDisposition = "recoverable"
	DenyTerminal    DenyDisposition = "terminal"
)

type GrantScope string

const (
	GrantOneShot GrantScope = "one_shot"
	GrantRun     GrantScope = "run"
)

// Decision describes policy output. It is not an execution permit and cannot
// be passed to Runtime.Invoke to authorize an effect.
type Decision struct {
	kind        DecisionKind
	reason      string
	disposition DenyDisposition
	scope       GrantScope
}

func NewDecision(kind DecisionKind, reason string, disposition DenyDisposition, scope GrantScope) (Decision, error) {
	if !safeCode(reason) {
		return Decision{}, fmt.Errorf("decision reason %q is not a safe code: %w", reason, ErrInvalidDecision)
	}
	switch kind {
	case DecisionAllow:
		if disposition != "" || (scope != GrantOneShot && scope != GrantRun) {
			return Decision{}, fmt.Errorf("allow requires a grant scope and no deny disposition: %w", ErrInvalidDecision)
		}
	case DecisionDeny:
		if scope != "" || (disposition != DenyRecoverable && disposition != DenyTerminal) {
			return Decision{}, fmt.Errorf("deny requires a disposition and no grant scope: %w", ErrInvalidDecision)
		}
	case DecisionPrompt:
		if disposition != "" || scope != "" {
			return Decision{}, fmt.Errorf("prompt cannot contain grant or deny state: %w", ErrInvalidDecision)
		}
	default:
		return Decision{}, fmt.Errorf("decision kind %q is invalid: %w", kind, ErrInvalidDecision)
	}
	return Decision{kind: kind, reason: reason, disposition: disposition, scope: scope}, nil
}

func (decision Decision) Kind() DecisionKind           { return decision.kind }
func (decision Decision) Reason() string               { return decision.reason }
func (decision Decision) Disposition() DenyDisposition { return decision.disposition }
func (decision Decision) Scope() GrantScope            { return decision.scope }
func (decision Decision) Validate() error {
	_, err := NewDecision(decision.kind, decision.reason, decision.disposition, decision.scope)
	return err
}

type ApprovalKind string

const (
	ApprovalAllow ApprovalKind = "allow"
	ApprovalDeny  ApprovalKind = "deny"
)

type Approval struct {
	kind        ApprovalKind
	reason      string
	disposition DenyDisposition
	scope       GrantScope
}

func NewApproval(kind ApprovalKind, reason string, disposition DenyDisposition, scope GrantScope) (Approval, error) {
	decisionKind := DecisionDeny
	if kind == ApprovalAllow {
		decisionKind = DecisionAllow
	} else if kind != ApprovalDeny {
		return Approval{}, ErrInvalidApproval
	}
	_, err := NewDecision(decisionKind, reason, disposition, scope)
	if err != nil {
		return Approval{}, fmt.Errorf("approval: %w: %w", err, ErrInvalidApproval)
	}
	return Approval{kind: kind, reason: reason, disposition: disposition, scope: scope}, nil
}

func (approval Approval) Kind() ApprovalKind           { return approval.kind }
func (approval Approval) Reason() string               { return approval.reason }
func (approval Approval) Disposition() DenyDisposition { return approval.disposition }
func (approval Approval) Scope() GrantScope            { return approval.scope }
func (approval Approval) Validate() error {
	_, err := NewApproval(approval.kind, approval.reason, approval.disposition, approval.scope)
	return err
}

type Policy interface {
	ID() PolicyID
	Decide(context.Context, PermissionRequest) (Decision, error)
}

type Approver interface {
	Approve(context.Context, PermissionRequest) (Approval, error)
}
