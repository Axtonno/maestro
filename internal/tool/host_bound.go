package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

var ErrRequestOutOfScope = errors.New("request_out_of_scope")

// HostBoundMutation is a trusted host API, never a model-facing tool. Capture
// must precede generation; only its returned immutable selection is accepted.
type HostBoundMutation struct {
	registry *WorkspaceRegistry
	tool     pkgTool.Tool
}

func NewHostBoundMutation(registry *WorkspaceRegistry) (*HostBoundMutation, error) {
	t, err := NewControlledMutationTool(registry)
	if err != nil {
		return nil, err
	}
	return &HostBoundMutation{registry, t}, nil
}

type BoundSelection struct {
	selection mutation.Selection
	run       pkgTool.RunID
	owner     *HostBoundMutation
}

func (s BoundSelection) Target() mutation.Selection { return s.selection }

func (h *HostBoundMutation) Capture(ctx context.Context, run pkgTool.RunID, paths []string, start, end int) (BoundSelection, error) {
	if len(paths) != 1 {
		return BoundSelection{}, ErrRequestOutOfScope
	}
	logical := paths[0]
	if err := mutation.ValidateTarget(logical); err != nil {
		return BoundSelection{}, mutation.ErrSensitiveTarget
	}
	if validateLogical(logical, false) != nil || !strings.HasPrefix(logical, "app/") || !strings.HasSuffix(logical, ".php") {
		return BoundSelection{}, mutation.ErrSensitiveTarget
	}
	w, ok := h.registry.Resolve(run)
	if !ok {
		return BoundSelection{}, pkgTool.ErrInvalidInvocation
	}
	root, err := openPhysicalRoot(w)
	if err != nil {
		return BoundSelection{}, err
	}
	defer root.Close()
	if err := validatePhysicalPath(root, logical, false, false); err != nil {
		return BoundSelection{}, err
	}
	content, err := readPhysicalFile(ctx, root, logical, w.Policy().MaxFileBytes)
	if err != nil {
		return BoundSelection{}, err
	}
	s, err := mutation.Select(mutation.Snapshot{Path: logical, Content: content, Digest: digest(content)}, start, end)
	if err != nil {
		return BoundSelection{}, err
	}
	return BoundSelection{s, run, h}, nil
}

func (h *HostBoundMutation) Prepare(ctx context.Context, bound BoundSelection, call pkgTool.CallID, raw []byte) (pkgTool.PreparedInvocation, error) {
	if bound.owner != h {
		return pkgTool.PreparedInvocation{}, pkgTool.ErrInvalidInvocation
	}
	decision, err := mutation.DecodeHostBoundDecision(raw)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	if decision.Decision == mutation.BinaryAbstain {
		return pkgTool.PreparedInvocation{}, mutation.ErrInsufficientInformation
	}
	s := bound.selection
	after, err := s.Replace(decision.NewText)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	// The legacy atomic engine receives a unique whole-file replacement made
	// by the host. Neither its target nor old text comes from the model.
	proposal, err := json.Marshal(mutation.ProposalV1{Version: 1, Path: s.Path(), Operation: "replace", OldText: s.Before(), NewText: after})
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	inv, err := pkgTool.NewInvocation(WorkspaceReplaceID, call, bound.run, proposal)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	p, err := h.tool.Prepare(ctx, inv)
	if err != nil {
		if errors.Is(err, mutation.ErrTargetNotFound) || errors.Is(err, mutation.ErrTargetAmbiguous) {
			return pkgTool.PreparedInvocation{}, mutation.ErrStaleSource
		}
		return pkgTool.PreparedInvocation{}, err
	}
	var args preparedReplaceArguments
	if err := json.Unmarshal(p.Arguments(), &args); err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	if args.ExpectedDigest != s.BeforeDigest() {
		return pkgTool.PreparedInvocation{}, mutation.ErrStaleSource
	}
	preview, ok := p.Preview()
	if !ok {
		return pkgTool.PreparedInvocation{}, pkgTool.ErrInvalidPreparedInvocation
	}
	args.Fingerprint = s.Fingerprint(decision.NewText, preview.Body())
	fields := []pkgTool.PreviewField{}
	for _, pair := range [][2]string{{"path", s.Path()}, {"start_line", fmt.Sprint(s.StartLine())}, {"end_line", fmt.Sprint(s.EndLine())}, {"before_sha256", s.BeforeDigest()}, {"selected_sha256", digest(s.Text())}, {"replacement_sha256", digest(decision.NewText)}, {"diff_sha256", digest(preview.Body())}, {"fingerprint", args.Fingerprint}} {
		field, err := pkgTool.NewPreviewField(pair[0], pair[1])
		if err != nil {
			return pkgTool.PreparedInvocation{}, err
		}
		fields = append(fields, field)
	}
	preview, err = pkgTool.NewPreview("Replace selected lines in "+s.Path(), fields, preview.Body(), "text/x-diff")
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	encoded, err := json.Marshal(args)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	return pkgTool.NewPreparedInvocationWithPreview(inv, h.tool.Descriptor().Version(), encoded, p.Actions(), preview)
}

func (h *HostBoundMutation) Execute(ctx context.Context, prepared pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	return h.tool.Execute(ctx, prepared)
}
