package llamacpp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

var sha256Digest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func (expected Attestation) configured() bool {
	return expected != (Attestation{})
}

func (expected Attestation) validate(baseURL, model string) error {
	if !expected.configured() {
		return nil
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme != "http" || !loopbackHost(parsed.Hostname()) {
		return fmt.Errorf("GGUF attestation requires a loopback HTTP server")
	}
	if model == "" {
		return fmt.Errorf("GGUF attestation requires a default model")
	}
	if expected.ModelPath == "" || !filepath.IsAbs(expected.ModelPath) ||
		filepath.Clean(expected.ModelPath) != expected.ModelPath {
		return fmt.Errorf("GGUF attestation model path must be absolute and normalized")
	}
	if !sha256Digest.MatchString(expected.ModelDigest) {
		return fmt.Errorf("GGUF attestation digest must be a lowercase SHA-256")
	}
	if strings.TrimSpace(expected.ServerBuild) == "" || strings.TrimSpace(expected.ServerBuild) != expected.ServerBuild {
		return fmt.Errorf("GGUF attestation requires an exact server build")
	}
	if expected.ContextWindow < 128 {
		return fmt.Errorf("GGUF attestation context window must be at least 128")
	}
	return nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (p *Provider) attestLocalModel(ctx context.Context, infos []pkgProvider.ModelInfo) ([]pkgProvider.ModelInfo, error) {
	if !p.attestation.configured() {
		return infos, nil
	}
	props := serverProps{}
	if err := p.doJSON(ctx, http.MethodGet, "/props", nil, &props); err != nil {
		return nil, fmt.Errorf("attest llama.cpp server: %w", err)
	}
	expected := p.attestation
	if props.ModelAlias != p.defaultModel || props.BuildInfo != expected.ServerBuild ||
		props.Settings.ContextWindow != expected.ContextWindow ||
		!samePath(props.ModelPath, expected.ModelPath) {
		return nil, fmt.Errorf("attest llama.cpp server identity: %w", pkgProvider.ErrInvalidResponse)
	}
	info, err := os.Lstat(expected.ModelPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("attest llama.cpp GGUF: model file unavailable: %w", pkgProvider.ErrInvalidResponse)
	}
	digest, err := hashRegularFile(expected.ModelPath, info)
	if err != nil {
		return nil, fmt.Errorf("attest llama.cpp GGUF: %w", err)
	}
	if digest != expected.ModelDigest {
		return nil, fmt.Errorf("attest llama.cpp GGUF digest: %w", pkgProvider.ErrInvalidResponse)
	}
	matched := false
	for index := range infos {
		if infos[index].Model.ID != p.defaultModel {
			continue
		}
		matched = true
		infos[index].Digest = digest
		infos[index].SizeBytes = info.Size()
		infos[index].ContextLength = expected.ContextWindow
		infos[index].Format = "gguf"
		infos[index].State = pkgProvider.ModelStateLoaded
	}
	if !matched {
		return nil, fmt.Errorf("attest llama.cpp model alias: %w", pkgProvider.ErrInvalidResponse)
	}
	return infos, nil
}

func samePath(left, right string) bool {
	leftAbsolute, leftErr := filepath.Abs(left)
	rightAbsolute, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	leftClean, rightClean := filepath.Clean(leftAbsolute), filepath.Clean(rightAbsolute)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(leftClean, rightClean)
	}
	return leftClean == rightClean
}

func hashRegularFile(path string, before os.FileInfo) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	magic := make([]byte, 4)
	if _, err := io.ReadFull(file, magic); err != nil || string(magic) != "GGUF" {
		_ = file.Close()
		return "", fmt.Errorf("model is not a GGUF file: %w", pkgProvider.ErrInvalidResponse)
	}
	_, _ = hash.Write(magic)
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() ||
		!before.ModTime().Equal(after.ModTime()) {
		return "", fmt.Errorf("GGUF changed during hashing: %w", pkgProvider.ErrInvalidResponse)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
