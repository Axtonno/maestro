package agent

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func planningPermission(request pkgAgent.RunRequest, bundle pkgContext.ContextBundle) (pkgTool.PermissionRequest, error) {
	target, err := pkgTool.NewModelTarget(request.Provider(), request.Model())
	if err != nil {
		return pkgTool.PermissionRequest{}, err
	}
	sections := bundle.Sections()
	if len(sections) == 0 {
		return pkgTool.NewModelPermissionRequest(request.Policy(), request.Run(), target, nil)
	}
	hash := sha256.New()
	totalBytes := int64(0)
	for _, section := range sections {
		writeDisclosurePart(hash, string(section.Path))
		writeDisclosurePart(hash, section.Text)
		totalBytes += int64(len(section.Text))
	}
	manifest, err := pkgTool.NewDisclosureManifest(
		bundle.Workspace(), bundle.Generation(), len(sections), bundle.UsedTokens(),
		totalBytes, pkgTool.Fingerprint(hex.EncodeToString(hash.Sum(nil))),
	)
	if err != nil {
		return pkgTool.PermissionRequest{}, err
	}
	return pkgTool.NewModelPermissionRequest(request.Policy(), request.Run(), target, &manifest)
}

type disclosureWriter interface{ Write([]byte) (int, error) }

func writeDisclosurePart(writer disclosureWriter, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write([]byte(value))
}
