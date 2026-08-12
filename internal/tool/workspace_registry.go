package tool

import (
	"fmt"
	"sync"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

// WorkspaceRegistry binds an immutable workspace root to an active run. It is
// shared by Agent Runtime and the reference workspace tools; roots never enter
// tool arguments or provider messages.
type WorkspaceRegistry struct {
	mu         sync.RWMutex
	workspaces map[pkgTool.RunID]pkgContext.Workspace
}

func NewWorkspaceRegistry() *WorkspaceRegistry {
	return &WorkspaceRegistry{workspaces: make(map[pkgTool.RunID]pkgContext.Workspace)}
}

func (registry *WorkspaceRegistry) Bind(run pkgTool.RunID, workspace pkgContext.Workspace) error {
	if err := run.Validate(); err != nil || workspace.Validate() != nil {
		return fmt.Errorf("workspace binding is invalid: %w", pkgTool.ErrInvalidInvocation)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.workspaces[run]; exists {
		return fmt.Errorf("workspace already bound for run %q: %w", run, pkgTool.ErrInvalidInvocation)
	}
	registry.workspaces[run] = workspace
	return nil
}

func (registry *WorkspaceRegistry) Unbind(run pkgTool.RunID) {
	registry.mu.Lock()
	delete(registry.workspaces, run)
	registry.mu.Unlock()
}

func (registry *WorkspaceRegistry) Resolve(run pkgTool.RunID) (pkgContext.Workspace, bool) {
	registry.mu.RLock()
	workspace, exists := registry.workspaces[run]
	registry.mu.RUnlock()
	return workspace, exists
}
