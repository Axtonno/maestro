package laravel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

const (
	ID      pkgPlugin.ID = "laravel"
	Version              = "0.3.0"

	maxComposerManifestBytes = 1 << 20
)

type Plugin interface {
	pkgPlugin.Plugin

	Root() string
	FrameworkVersion() string
	pkgContext.WorkspaceProvider
}

var _ Plugin = (*plugin)(nil)
var _ pkgRuntime.Initializer = (*plugin)(nil)
var _ pkgRuntime.HealthChecker = (*plugin)(nil)
var _ pkgContext.WorkspaceProvider = (*plugin)(nil)

type plugin struct {
	mu sync.RWMutex

	root             string
	frameworkVersion string
}

func New(root string) (Plugin, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf(
			"create Laravel plugin: workspace root is empty: %w",
			ErrInvalidConfig,
		)
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf(
			"create Laravel plugin: resolve workspace root %q: %w",
			root,
			errors.Join(ErrInvalidConfig, err),
		)
	}

	return &plugin{root: filepath.Clean(absoluteRoot)}, nil
}

func (p *plugin) Metadata() pkgRuntime.Metadata {
	return pkgRuntime.Metadata{
		ID:          ID,
		Name:        "Laravel",
		Version:     Version,
		Description: "Laravel framework workspace integration",
		Capabilities: []pkgRuntime.Capability{
			pkgRuntime.CapabilityInitialize,
			pkgRuntime.CapabilityHealth,
			pkgPlugin.CapabilityWorkspaceDetection,
			pkgContext.CapabilityWorkspaceProvider,
		},
	}
}

func (p *plugin) Manifest() pkgPlugin.Manifest {
	return pkgPlugin.Manifest{
		RuntimeAPIVersion: pkgPlugin.RuntimeAPIVersion,
	}
}

func (p *plugin) Initialize(pkgRuntime.Context) error {
	frameworkVersion, err := detect(p.root)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.frameworkVersion = frameworkVersion
	p.mu.Unlock()

	return nil
}

func (p *plugin) Health(pkgRuntime.Context) error {
	_, err := detect(p.root)

	return err
}

func (p *plugin) Root() string {
	return p.root
}

func (p *plugin) FrameworkVersion() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.frameworkVersion
}

func (p *plugin) Workspace(ctx context.Context) (pkgContext.Workspace, error) {
	if ctx == nil {
		return pkgContext.Workspace{}, fmt.Errorf("describe Laravel workspace with nil context: %w", pkgContext.ErrInvalidWorkspace)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.Workspace{}, err
	}
	metadata := map[string]string{"framework": "laravel"}
	if version := p.FrameworkVersion(); version != "" {
		metadata["framework_version"] = version
	}
	return pkgContext.NewWorkspace(
		pkgContext.WorkspaceID(ID),
		p.root,
		pkgContext.WorkspaceOptions{
			Source:   pkgContext.SourceFilesystem,
			Policy:   pkgContext.DefaultScanPolicy(),
			Metadata: metadata,
		},
	)
}

func detect(root string) (string, error) {
	artisanPath := filepath.Join(root, "artisan")
	artisanInfo, err := os.Stat(artisanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"detect Laravel at %q: artisan is missing: %w",
				root,
				ErrNotDetected,
			)
		}

		return "", fmt.Errorf(
			"detect Laravel at %q: inspect artisan: %w",
			root,
			err,
		)
	}
	if !artisanInfo.Mode().IsRegular() {
		return "", fmt.Errorf(
			"detect Laravel at %q: artisan is not a regular file: %w",
			root,
			ErrNotDetected,
		)
	}

	composerPath := filepath.Join(root, "composer.json")
	manifest, err := readComposerManifest(composerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"detect Laravel at %q: composer.json is missing: %w",
				root,
				ErrNotDetected,
			)
		}

		return "", err
	}

	frameworkVersion := strings.TrimSpace(manifest.Require["laravel/framework"])
	if frameworkVersion == "" {
		return "", fmt.Errorf(
			"detect Laravel at %q: composer.json does not require laravel/framework: %w",
			root,
			ErrNotDetected,
		)
	}

	return frameworkVersion, nil
}

type composerManifest struct {
	Require map[string]string `json:"require"`
}

func readComposerManifest(path string) (composerManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return composerManifest{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(
		file,
		maxComposerManifestBytes+1,
	))
	if err != nil {
		return composerManifest{}, fmt.Errorf(
			"read Composer manifest %q: %w",
			path,
			err,
		)
	}
	if len(content) > maxComposerManifestBytes {
		return composerManifest{}, fmt.Errorf(
			"read Composer manifest %q: manifest exceeds %d bytes: %w",
			path,
			maxComposerManifestBytes,
			ErrInvalidComposerManifest,
		)
	}

	var manifest composerManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return composerManifest{}, fmt.Errorf(
			"parse Composer manifest %q: %w",
			path,
			errors.Join(ErrInvalidComposerManifest, err),
		)
	}

	return manifest, nil
}
