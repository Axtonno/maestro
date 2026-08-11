package contextengine

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const SourceFilesystem SourceID = "context.filesystem"

type ScanPolicy struct {
	MaxFiles      int
	MaxFileBytes  int64
	MaxTotalBytes int64
	Include       []string
	Exclude       []string
	IncludeHidden bool
	IncludeBinary bool
}

func DefaultScanPolicy() ScanPolicy {
	return ScanPolicy{
		MaxFiles:      10_000,
		MaxFileBytes:  1 << 20,
		MaxTotalBytes: 64 << 20,
		Exclude: []string{
			".git", ".git/**", "node_modules", "node_modules/**",
			"vendor", "vendor/**",
		},
	}
}

func (policy ScanPolicy) Validate() error {
	if policy.MaxFiles <= 0 || policy.MaxFileBytes <= 0 || policy.MaxTotalBytes <= 0 ||
		policy.MaxFileBytes > policy.MaxTotalBytes {
		return fmt.Errorf("scan limits must be positive and per-file must not exceed total: %w", ErrInvalidPolicy)
	}
	for _, group := range [][]string{policy.Include, policy.Exclude} {
		for index, pattern := range group {
			if pattern == "" || strings.TrimSpace(pattern) != pattern || strings.Contains(pattern, "\\") {
				return fmt.Errorf("scan pattern %d %q is not normalized: %w", index, pattern, ErrInvalidPolicy)
			}
			if _, err := path.Match(pattern, "probe"); err != nil {
				return fmt.Errorf("scan pattern %d %q: %w: %w", index, pattern, err, ErrInvalidPolicy)
			}
		}
	}
	return nil
}

type WorkspaceOptions struct {
	Source   SourceID
	Policy   ScanPolicy
	Metadata map[string]string
}

// Workspace is immutable after construction. Metadata and policy patterns are
// returned as defensive copies.
type Workspace struct {
	id       WorkspaceID
	root     string
	source   SourceID
	policy   ScanPolicy
	metadata map[string]string
}

func NewWorkspace(id WorkspaceID, root string, options WorkspaceOptions) (Workspace, error) {
	if err := id.Validate(); err != nil {
		return Workspace{}, fmt.Errorf("workspace identity: %w: %w", err, ErrInvalidWorkspace)
	}
	if root == "" || !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return Workspace{}, fmt.Errorf("workspace root %q must be absolute and normalized: %w", root, ErrInvalidWorkspace)
	}
	if err := options.Source.Validate(); err != nil {
		return Workspace{}, fmt.Errorf("workspace source: %w: %w", err, ErrInvalidWorkspace)
	}
	if err := options.Policy.Validate(); err != nil {
		return Workspace{}, fmt.Errorf("workspace policy: %w: %w", err, ErrInvalidWorkspace)
	}
	metadata := make(map[string]string, len(options.Metadata))
	for key, value := range options.Metadata {
		if !exactID(key) || strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return Workspace{}, fmt.Errorf("workspace metadata key %q is not safe: %w", key, ErrInvalidWorkspace)
		}
		metadata[key] = value
	}
	return Workspace{id: id, root: root, source: options.Source, policy: clonePolicy(options.Policy), metadata: metadata}, nil
}

func (workspace Workspace) ID() WorkspaceID    { return workspace.id }
func (workspace Workspace) Root() string       { return workspace.root }
func (workspace Workspace) Source() SourceID   { return workspace.source }
func (workspace Workspace) Policy() ScanPolicy { return clonePolicy(workspace.policy) }
func (workspace Workspace) Metadata() map[string]string {
	metadata := make(map[string]string, len(workspace.metadata))
	for key, value := range workspace.metadata {
		metadata[key] = value
	}
	return metadata
}

func (workspace Workspace) Validate() error {
	_, err := NewWorkspace(workspace.id, workspace.root, WorkspaceOptions{
		Source: workspace.source, Policy: workspace.policy, Metadata: workspace.metadata,
	})
	return err
}

type WorkspaceProvider interface {
	Workspace(context.Context) (Workspace, error)
}

func clonePolicy(policy ScanPolicy) ScanPolicy {
	policy.Include = slices.Clone(policy.Include)
	policy.Exclude = slices.Clone(policy.Exclude)
	return policy
}
