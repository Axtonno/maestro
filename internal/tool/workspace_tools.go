package tool

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const (
	WorkspaceListID   pkgTool.ID = "workspace.list"
	WorkspaceReadID   pkgTool.ID = "workspace.read"
	WorkspaceSearchID pkgTool.ID = "workspace.search"
	WorkspaceWriteID  pkgTool.ID = "workspace.write"
	WorkspacePatchID  pkgTool.ID = "workspace.patch"

	maxPatchPreviewBytes  = 256 << 10
	patchDiffContextLines = 3
)

type workspaceOperation string

const (
	workspaceList   workspaceOperation = "list"
	workspaceRead   workspaceOperation = "read"
	workspaceSearch workspaceOperation = "search"
	workspaceWrite  workspaceOperation = "write"
	workspacePatch  workspaceOperation = "patch"
)

type workspaceTool struct {
	registry   *WorkspaceRegistry
	operation  workspaceOperation
	descriptor pkgTool.Descriptor
	atomicOps  atomicFileOps
}

// NewWorkspaceTools constructs the framework-neutral reference filesystem
// tool set. Callers register the returned trusted in-process tools explicitly.
func NewWorkspaceTools(registry *WorkspaceRegistry) ([]pkgTool.Tool, error) {
	if registry == nil {
		return nil, fmt.Errorf("workspace registry is nil: %w", pkgTool.ErrInvalidTool)
	}
	specifications := []struct {
		operation   workspaceOperation
		id          pkgTool.ID
		name        pkgTool.Name
		description string
		schema      string
		effect      pkgTool.Effect
	}{
		{workspaceList, WorkspaceListID, "workspace_list", "List physical entries beneath one logical workspace directory.", `{"type":"object","additionalProperties":false,"properties":{"path":{"type":"string"},"max_entries":{"type":"integer","minimum":1,"maximum":1000}}}`, pkgTool.EffectWorkspaceInspect},
		{workspaceRead, WorkspaceReadID, "workspace_read", "Read one physical UTF-8 file and return its content digest.", `{"type":"object","additionalProperties":false,"required":["path"],"properties":{"path":{"type":"string"},"max_bytes":{"type":"integer","minimum":1}}}`, pkgTool.EffectWorkspaceInspect},
		{workspaceSearch, WorkspaceSearchID, "workspace_search", "Search UTF-8 files below a logical workspace directory.", `{"type":"object","additionalProperties":false,"required":["query"],"properties":{"query":{"type":"string"},"path":{"type":"string"},"max_results":{"type":"integer","minimum":1,"maximum":1000}}}`, pkgTool.EffectWorkspaceInspect},
		{workspaceWrite, WorkspaceWriteID, "workspace_write", "Write one UTF-8 file when its expected digest still matches.", `{"type":"object","additionalProperties":false,"required":["path","content","expected_digest"],"properties":{"path":{"type":"string"},"content":{"type":"string"},"expected_digest":{"type":"string"}}}`, pkgTool.EffectWorkspaceMutate},
		{workspacePatch, WorkspacePatchID, "workspace_patch", "Replace one exact text occurrence when the file digest still matches.", `{"type":"object","additionalProperties":false,"required":["path","old","new","expected_digest"],"properties":{"path":{"type":"string"},"old":{"type":"string"},"new":{"type":"string"},"expected_digest":{"type":"string"}}}`, pkgTool.EffectWorkspaceMutate},
	}
	tools := make([]pkgTool.Tool, 0, len(specifications))
	for _, specification := range specifications {
		descriptor, err := pkgTool.NewDescriptor(
			specification.id, specification.name, "1.0.0", specification.description,
			json.RawMessage(specification.schema), []pkgTool.Effect{specification.effect},
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, &workspaceTool{
			registry: registry, operation: specification.operation,
			descriptor: descriptor, atomicOps: defaultAtomicFileOps(),
		})
	}
	return tools, nil
}

func (tool *workspaceTool) Descriptor() pkgTool.Descriptor { return tool.descriptor }

func (tool *workspaceTool) Prepare(ctx context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
	workspace, exists := tool.registry.Resolve(invocation.Run())
	if !exists {
		return pkgTool.PreparedInvocation{}, fmt.Errorf("workspace is not bound to run: %w", pkgTool.ErrInvalidInvocation)
	}
	normalized, logical, err := tool.normalize(invocation.Arguments(), workspace)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	effect := pkgTool.EffectWorkspaceInspect
	if tool.operation == workspaceWrite || tool.operation == workspacePatch {
		effect = pkgTool.EffectWorkspaceMutate
	}
	action, err := pkgTool.NewAction(effect, logical, workspace.ID())
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	if tool.operation == workspacePatch {
		return tool.preparePatch(ctx, invocation, workspace, normalized, action)
	}
	return pkgTool.NewPreparedInvocation(invocation, tool.descriptor.Version(), normalized, []pkgTool.Action{action})
}

func (tool *workspaceTool) Execute(ctx context.Context, prepared pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	workspace, exists := tool.registry.Resolve(prepared.Invocation().Run())
	if !exists {
		return pkgTool.Result{}, fmt.Errorf("workspace binding disappeared: %w", pkgTool.ErrExecutionFailed)
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Result{}, err
	}
	root, err := openPhysicalRoot(workspace)
	if err != nil {
		return pkgTool.Result{}, err
	}
	defer root.Close()
	switch tool.operation {
	case workspaceList:
		return executeList(ctx, root, prepared.Arguments())
	case workspaceRead:
		return executeRead(ctx, root, workspace, prepared.Arguments())
	case workspaceSearch:
		return executeSearch(ctx, root, workspace, prepared.Arguments())
	case workspaceWrite:
		return executeWrite(ctx, root, workspace, prepared.Arguments())
	case workspacePatch:
		return executePatch(ctx, root, workspace, prepared.Arguments(), tool.atomicOps)
	default:
		return pkgTool.Result{}, pkgTool.ErrExecutionFailed
	}
}

type listArguments struct {
	Path       string `json:"path,omitempty"`
	MaxEntries int    `json:"max_entries,omitempty"`
}

type readArguments struct {
	Path     string `json:"path"`
	MaxBytes int64  `json:"max_bytes,omitempty"`
}

type searchArguments struct {
	Query      string `json:"query"`
	Path       string `json:"path,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

type writeArguments struct {
	Path           string `json:"path"`
	Content        string `json:"content"`
	ExpectedDigest string `json:"expected_digest"`
}

type patchArguments struct {
	Path           string `json:"path"`
	Old            string `json:"old"`
	New            string `json:"new"`
	ExpectedDigest string `json:"expected_digest"`
}

type preparedPatchArguments struct {
	Path            string `json:"path"`
	Old             string `json:"old"`
	New             string `json:"new"`
	ExpectedDigest  string `json:"expected_digest"`
	ProposedContent string `json:"proposed_content"`
}

func (tool *workspaceTool) preparePatch(
	ctx context.Context,
	invocation pkgTool.Invocation,
	workspace pkgContext.Workspace,
	normalized json.RawMessage,
	action pkgTool.Action,
) (pkgTool.PreparedInvocation, error) {
	if err := ctx.Err(); err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	var arguments patchArguments
	if err := decodeStrict(normalized, &arguments); err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	if err := validateSupportedPatchPath(arguments.Path); err != nil {
		return pkgTool.PreparedInvocation{}, pkgTool.ErrInvalidInvocation
	}
	root, err := openPhysicalRoot(workspace)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	defer root.Close()
	if err := validatePhysicalPath(root, arguments.Path, false, false); err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	content, err := readPhysicalFile(ctx, root, arguments.Path, workspace.Policy().MaxFileBytes)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	if digest(content) != arguments.ExpectedDigest || strings.Count(content, arguments.Old) != 1 || arguments.Old == arguments.New {
		return pkgTool.PreparedInvocation{}, fmt.Errorf("patch proposal precondition failed: %w", pkgTool.ErrInvalidInvocation)
	}
	proposed := strings.Replace(content, arguments.Old, arguments.New, 1)
	if int64(len(proposed)) > workspace.Policy().MaxFileBytes {
		return pkgTool.PreparedInvocation{}, pkgTool.ErrLimitExceeded
	}
	diff, err := renderPatchDiff(arguments.Path, content, proposed)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	fields := make([]pkgTool.PreviewField, 0, 4)
	for _, value := range [][2]string{
		{"tool", string(WorkspacePatchID)},
		{"path", arguments.Path},
		{"expected_sha256", arguments.ExpectedDigest},
		{"precondition", "existing_regular_utf8_php_single_exact_occurrence"},
	} {
		field, fieldErr := pkgTool.NewPreviewField(value[0], value[1])
		if fieldErr != nil {
			return pkgTool.PreparedInvocation{}, fieldErr
		}
		fields = append(fields, field)
	}
	preview, err := pkgTool.NewPreview(
		"Replace one exact occurrence in "+arguments.Path,
		fields,
		diff,
		"text/x-diff",
	)
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	preparedArguments, err := json.Marshal(preparedPatchArguments{
		Path: arguments.Path, Old: arguments.Old, New: arguments.New,
		ExpectedDigest: arguments.ExpectedDigest, ProposedContent: proposed,
	})
	if err != nil {
		return pkgTool.PreparedInvocation{}, err
	}
	return pkgTool.NewPreparedInvocationWithPreview(
		invocation, tool.descriptor.Version(), preparedArguments,
		[]pkgTool.Action{action}, preview,
	)
}

func (tool *workspaceTool) normalize(arguments json.RawMessage, workspace pkgContext.Workspace) (json.RawMessage, string, error) {
	switch tool.operation {
	case workspaceList:
		var value listArguments
		if err := decodeStrict(arguments, &value); err != nil {
			return nil, "", err
		}
		value.Path = normalizeDirectory(value.Path)
		if value.MaxEntries == 0 {
			value.MaxEntries = 200
		}
		if value.MaxEntries < 1 || value.MaxEntries > 1000 || validateLogical(value.Path, true) != nil {
			return nil, "", pkgTool.ErrInvalidInvocation
		}
		return marshalArguments(value, value.Path)
	case workspaceRead:
		var value readArguments
		if err := decodeStrict(arguments, &value); err != nil {
			return nil, "", err
		}
		if value.MaxBytes == 0 {
			value.MaxBytes = workspace.Policy().MaxFileBytes
		}
		if validateLogical(value.Path, false) != nil || value.MaxBytes < 1 || value.MaxBytes > workspace.Policy().MaxFileBytes {
			return nil, "", pkgTool.ErrInvalidInvocation
		}
		return marshalArguments(value, value.Path)
	case workspaceSearch:
		var value searchArguments
		if err := decodeStrict(arguments, &value); err != nil {
			return nil, "", err
		}
		value.Path = normalizeDirectory(value.Path)
		if value.MaxResults == 0 {
			value.MaxResults = 100
		}
		if validateLogical(value.Path, true) != nil || strings.TrimSpace(value.Query) == "" || len(value.Query) > 4096 ||
			!utf8.ValidString(value.Query) || strings.ContainsRune(value.Query, 0) || value.MaxResults < 1 || value.MaxResults > 1000 {
			return nil, "", pkgTool.ErrInvalidInvocation
		}
		return marshalArguments(value, value.Path)
	case workspaceWrite:
		var value writeArguments
		if err := decodeStrict(arguments, &value); err != nil {
			return nil, "", err
		}
		if validateMutation(value.Path, value.Content, value.ExpectedDigest, workspace) != nil {
			return nil, "", pkgTool.ErrInvalidInvocation
		}
		return marshalArguments(value, value.Path)
	case workspacePatch:
		var value patchArguments
		if err := decodeStrict(arguments, &value); err != nil {
			return nil, "", err
		}
		if validateLogical(value.Path, false) != nil || value.Old == "" || len(value.Old)+len(value.New) > int(workspace.Policy().MaxFileBytes) ||
			!utf8.ValidString(value.Old) || !utf8.ValidString(value.New) || strings.ContainsRune(value.Old, 0) || strings.ContainsRune(value.New, 0) ||
			validateExpectedDigest(value.ExpectedDigest, false) != nil {
			return nil, "", pkgTool.ErrInvalidInvocation
		}
		return marshalArguments(value, value.Path)
	default:
		return nil, "", pkgTool.ErrInvalidInvocation
	}
}

func validateSupportedPatchPath(logical string) error {
	if validateLogical(logical, false) != nil || !strings.HasPrefix(logical, "app/") || !strings.HasSuffix(logical, ".php") {
		return pkgContext.ErrInvalidPath
	}
	for _, part := range strings.Split(logical, "/") {
		if strings.HasPrefix(part, ".") {
			return pkgContext.ErrInvalidPath
		}
	}
	return nil
}

func decodeStrict(arguments json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode workspace tool arguments: %w: %w", err, pkgTool.ErrInvalidInvocation)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("workspace tool arguments contain trailing JSON: %w", pkgTool.ErrInvalidInvocation)
	}
	return nil
}

func marshalArguments(value any, logical string) (json.RawMessage, string, error) {
	normalized, err := json.Marshal(value)
	if err != nil {
		return nil, "", err
	}
	return normalized, logical, nil
}

func normalizeDirectory(value string) string {
	if value == "" {
		return "."
	}
	return value
}

func validateLogical(value string, allowRoot bool) error {
	if allowRoot && value == "." {
		return nil
	}
	return pkgContext.DocumentPath(value).Validate()
}

func validateMutation(logical, content, expected string, workspace pkgContext.Workspace) error {
	if validateLogical(logical, false) != nil || !utf8.ValidString(content) || strings.ContainsRune(content, 0) ||
		int64(len(content)) > workspace.Policy().MaxFileBytes {
		return pkgTool.ErrInvalidInvocation
	}
	return validateExpectedDigest(expected, true)
}

func validateExpectedDigest(value string, allowAbsent bool) error {
	if allowAbsent && value == "absent" {
		return nil
	}
	return pkgContext.Digest(value).Validate()
}

func openPhysicalRoot(workspace pkgContext.Workspace) (*os.Root, error) {
	info, err := os.Lstat(workspace.Root())
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, pkgContext.ErrInvalidWorkspace
	}
	return os.OpenRoot(workspace.Root())
}

func validatePhysicalPath(root *os.Root, logical string, allowRoot, allowMissingFinal bool) error {
	if err := validateLogical(logical, allowRoot); err != nil {
		return err
	}
	if logical == "." {
		return nil
	}
	parts := strings.Split(logical, "/")
	for index := range parts {
		candidate := path.Join(parts[:index+1]...)
		info, err := root.Lstat(candidate)
		if errors.Is(err, fs.ErrNotExist) && allowMissingFinal && index == len(parts)-1 {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("logical path %q contains a symlink: %w", logical, pkgContext.ErrInvalidPath)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("logical path %q has a non-directory parent: %w", logical, pkgContext.ErrInvalidPath)
		}
	}
	return nil
}

func executeList(ctx context.Context, root *os.Root, raw json.RawMessage) (pkgTool.Result, error) {
	var arguments listArguments
	if err := decodeStrict(raw, &arguments); err != nil {
		return pkgTool.Result{}, err
	}
	if err := validatePhysicalPath(root, arguments.Path, true, false); err != nil {
		return pkgTool.Result{}, err
	}
	directory, err := root.Open(arguments.Path)
	if err != nil {
		return pkgTool.Result{}, err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(arguments.MaxEntries + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return pkgTool.Result{}, err
	}
	truncated := len(entries) > arguments.MaxEntries
	if truncated {
		entries = entries[:arguments.MaxEntries]
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() < entries[right].Name() })
	type entry struct {
		Path string `json:"path"`
		Type string `json:"type"`
		Size int64  `json:"size,omitempty"`
	}
	results := make([]entry, 0, len(entries))
	for _, item := range entries {
		if err := ctx.Err(); err != nil {
			return pkgTool.Result{}, err
		}
		if item.Type()&os.ModeSymlink != 0 {
			continue
		}
		kind := "file"
		if item.IsDir() {
			kind = "directory"
		}
		info, err := item.Info()
		if err != nil {
			return pkgTool.Result{}, err
		}
		logical := item.Name()
		if arguments.Path != "." {
			logical = path.Join(arguments.Path, item.Name())
		}
		results = append(results, entry{Path: logical, Type: kind, Size: info.Size()})
	}
	encoded, _ := json.Marshal(results)
	return pkgTool.NewResult(pkgTool.ResultSuccess, string(encoded), "application/json", "listed", len(results), truncated, "")
}

func executeRead(ctx context.Context, root *os.Root, workspace pkgContext.Workspace, raw json.RawMessage) (pkgTool.Result, error) {
	var arguments readArguments
	if err := decodeStrict(raw, &arguments); err != nil {
		return pkgTool.Result{}, err
	}
	if err := validatePhysicalPath(root, arguments.Path, false, false); err != nil {
		return pkgTool.Result{}, err
	}
	content, err := readPhysicalFile(ctx, root, arguments.Path, min64(arguments.MaxBytes, workspace.Policy().MaxFileBytes))
	if err != nil {
		return pkgTool.Result{}, err
	}
	payload := struct {
		Path    string `json:"path"`
		Digest  string `json:"digest"`
		Content string `json:"content"`
	}{arguments.Path, digest(content), content}
	encoded, _ := json.Marshal(payload)
	return pkgTool.NewResult(pkgTool.ResultSuccess, string(encoded), "application/json", "read", 1, false, "")
}

func executeSearch(ctx context.Context, root *os.Root, workspace pkgContext.Workspace, raw json.RawMessage) (pkgTool.Result, error) {
	var arguments searchArguments
	if err := decodeStrict(raw, &arguments); err != nil {
		return pkgTool.Result{}, err
	}
	if err := validatePhysicalPath(root, arguments.Path, true, false); err != nil {
		return pkgTool.Result{}, err
	}
	type match struct {
		Path string `json:"path"`
		Line int    `json:"line"`
		Text string `json:"text"`
	}
	matches := make([]match, 0)
	err := fs.WalkDir(root.FS(), arguments.Path, func(logical string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if logical == "." || entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if len(matches) >= arguments.MaxResults {
			return fs.SkipAll
		}
		if err := validatePhysicalPath(root, logical, false, false); err != nil {
			return err
		}
		content, err := readPhysicalFile(ctx, root, logical, workspace.Policy().MaxFileBytes)
		if err != nil {
			return err
		}
		for index, line := range strings.Split(content, "\n") {
			if strings.Contains(line, arguments.Query) {
				matches = append(matches, match{Path: logical, Line: index + 1, Text: line})
				if len(matches) >= arguments.MaxResults {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return pkgTool.Result{}, err
	}
	encoded, _ := json.Marshal(matches)
	return pkgTool.NewResult(pkgTool.ResultSuccess, string(encoded), "application/json", "searched", len(matches), len(matches) >= arguments.MaxResults, "")
}

func executeWrite(ctx context.Context, root *os.Root, workspace pkgContext.Workspace, raw json.RawMessage) (pkgTool.Result, error) {
	var arguments writeArguments
	if err := decodeStrict(raw, &arguments); err != nil {
		return pkgTool.Result{}, err
	}
	if err := validatePhysicalPath(root, arguments.Path, false, arguments.ExpectedDigest == "absent"); err != nil {
		return pkgTool.Result{}, err
	}
	if arguments.ExpectedDigest == "absent" {
		if err := createPhysicalFile(ctx, root, arguments.Path, arguments.Content); err != nil {
			if errors.Is(err, fs.ErrExist) {
				return pkgTool.NewResult(pkgTool.ResultFailed, "", "", "precondition_failed", 0, false, "")
			}
			return pkgTool.Result{}, err
		}
	} else {
		matched, err := replacePhysicalFile(ctx, root, arguments.Path, arguments.ExpectedDigest, workspace.Policy().MaxFileBytes, func(string) (string, bool) {
			return arguments.Content, true
		})
		if err != nil {
			return pkgTool.Result{}, err
		}
		if !matched {
			return pkgTool.NewResult(pkgTool.ResultFailed, "", "", "precondition_failed", 0, false, "")
		}
	}
	return mutationResult(arguments.Path, arguments.Content)
}

func executePatch(
	ctx context.Context,
	root *os.Root,
	workspace pkgContext.Workspace,
	raw json.RawMessage,
	ops atomicFileOps,
) (pkgTool.Result, error) {
	var arguments preparedPatchArguments
	if err := decodeStrict(raw, &arguments); err != nil {
		return pkgTool.Result{}, err
	}
	if err := validatePhysicalPath(root, arguments.Path, false, false); err != nil {
		return pkgTool.Result{}, err
	}
	if int64(len(arguments.ProposedContent)) > workspace.Policy().MaxFileBytes {
		return pkgTool.Result{}, pkgTool.ErrInvalidPreparedInvocation
	}
	outcome, err := replacePhysicalFileAtomically(
		ctx, root, arguments.Path, arguments.ExpectedDigest,
		arguments.Old, arguments.New, arguments.ProposedContent,
		workspace.Policy().MaxFileBytes, ops,
	)
	if outcome.committed {
		resultOutcome := pkgTool.ResultSuccess
		reason := "written"
		if err != nil {
			resultOutcome = pkgTool.ResultFailed
			reason = "post_commit_sync_failed"
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				resultOutcome = pkgTool.ResultCanceled
				reason = "post_commit_canceled"
			}
		}
		return pkgTool.NewEffectResult(
			resultOutcome,
			atomicResultContent(arguments.Path, arguments.ProposedContent, true, outcome.durable),
			"application/json", reason, 1, false, "", pkgTool.EffectApplied, outcome.durable,
		)
	}
	if err != nil {
		return pkgTool.Result{}, err
	}
	if !outcome.matched {
		return pkgTool.NewEffectResult(pkgTool.ResultFailed, "", "", "precondition_failed", 0, false, "", pkgTool.EffectUnchanged, false)
	}
	return pkgTool.Result{}, pkgTool.ErrExecutionFailed
}

func renderPatchDiff(logical, before, after string) (string, error) {
	oldLines := splitDiffLines(before)
	newLines := splitDiffLines(after)
	prefix := 0
	for prefix < len(oldLines) && prefix < len(newLines) && oldLines[prefix] == newLines[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(oldLines)-prefix && suffix < len(newLines)-prefix &&
		oldLines[len(oldLines)-1-suffix] == newLines[len(newLines)-1-suffix] {
		suffix++
	}
	if prefix == len(oldLines) && prefix == len(newLines) {
		return "", fmt.Errorf("patch preview has no changes: %w", pkgTool.ErrInvalidInvocation)
	}
	start := max(0, prefix-patchDiffContextLines)
	oldChangeEnd := len(oldLines) - suffix
	newChangeEnd := len(newLines) - suffix
	oldEnd := min(len(oldLines), oldChangeEnd+patchDiffContextLines)
	newEnd := min(len(newLines), newChangeEnd+patchDiffContextLines)

	var rendered strings.Builder
	fmt.Fprintf(&rendered, "--- a/%s\n+++ b/%s\n", logical, logical)
	fmt.Fprintf(&rendered, "@@ -%d,%d +%d,%d @@\n", start+1, oldEnd-start, start+1, newEnd-start)
	for _, line := range oldLines[start:prefix] {
		appendDiffLine(&rendered, ' ', line)
	}
	for _, line := range oldLines[prefix:oldChangeEnd] {
		appendDiffLine(&rendered, '-', line)
	}
	for _, line := range newLines[prefix:newChangeEnd] {
		appendDiffLine(&rendered, '+', line)
	}
	oldContextStart := oldChangeEnd
	newContextStart := newChangeEnd
	contextLines := min(oldEnd-oldContextStart, newEnd-newContextStart)
	for index := 0; index < contextLines; index++ {
		appendDiffLine(&rendered, ' ', oldLines[oldContextStart+index])
	}
	if rendered.Len() > maxPatchPreviewBytes {
		return "", fmt.Errorf("patch preview exceeds %d bytes: %w", maxPatchPreviewBytes, pkgTool.ErrLimitExceeded)
	}
	return rendered.String(), nil
}

func splitDiffLines(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func appendDiffLine(builder *strings.Builder, marker byte, line string) {
	builder.WriteByte(marker)
	builder.WriteString(line)
	if !strings.HasSuffix(line, "\n") {
		builder.WriteString("\n\\ No newline at end of file\n")
	}
}

func readPhysicalFile(ctx context.Context, root *os.Root, logical string, maxBytes int64) (string, error) {
	file, err := root.Open(logical)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.Join(err, pkgContext.ErrInvalidPath)
	}
	if info.Size() > maxBytes {
		return "", pkgTool.ErrLimitExceeded
	}
	content, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if int64(len(content)) > maxBytes || !utf8.Valid(content) || bytes.IndexByte(content, 0) >= 0 {
		return "", pkgTool.ErrLimitExceeded
	}
	return string(content), nil
}

func createPhysicalFile(ctx context.Context, root *os.Root, logical, content string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := root.OpenFile(logical, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := io.WriteString(file, content)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.Join(writeErr, closeErr)
	}
	return ctx.Err()
}

func replacePhysicalFile(
	ctx context.Context,
	root *os.Root,
	logical, expected string,
	maxBytes int64,
	transform func(string) (string, bool),
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	file, err := root.OpenFile(logical, os.O_RDWR, 0)
	if err != nil {
		return false, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false, errors.Join(err, pkgContext.ErrInvalidPath)
	}
	if info.Size() > maxBytes {
		return false, pkgTool.ErrLimitExceeded
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if int64(len(data)) > maxBytes || !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 || digest(string(data)) != expected {
		return false, nil
	}
	updated, ok := transform(string(data))
	if !ok {
		return false, nil
	}
	if err := file.Truncate(0); err != nil {
		return false, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	if _, err := io.WriteString(file, updated); err != nil {
		return false, err
	}
	if err := file.Sync(); err != nil {
		return false, err
	}
	return true, ctx.Err()
}

func mutationResult(logical, content string) (pkgTool.Result, error) {
	encoded, _ := json.Marshal(struct {
		Path   string `json:"path"`
		Digest string `json:"digest"`
	}{logical, digest(content)})
	return pkgTool.NewResult(pkgTool.ResultSuccess, string(encoded), "application/json", "written", 1, false, "")
}

func digest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func min64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}
