package benchmark

import (
	"fmt"
	"io"
	"os"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	"gopkg.in/yaml.v3"
)

func LoadManifest(path string) (pkgBenchmark.Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return pkgBenchmark.Manifest{}, fmt.Errorf(
			"open benchmark manifest: %w",
			err,
		)
	}
	defer file.Close()

	manifest, err := DecodeManifest(file)
	if err != nil {
		return pkgBenchmark.Manifest{}, fmt.Errorf(
			"decode benchmark manifest %q: %w",
			path,
			err,
		)
	}

	return manifest, nil
}

func DecodeManifest(reader io.Reader) (pkgBenchmark.Manifest, error) {
	if reader == nil {
		return pkgBenchmark.Manifest{}, fmt.Errorf(
			"manifest reader is nil: %w",
			pkgBenchmark.ErrInvalidManifest,
		)
	}

	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	var manifest pkgBenchmark.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return pkgBenchmark.Manifest{}, fmt.Errorf("decode YAML: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return pkgBenchmark.Manifest{}, fmt.Errorf(
				"manifest contains multiple YAML documents: %w",
				pkgBenchmark.ErrInvalidManifest,
			)
		}
		return pkgBenchmark.Manifest{}, fmt.Errorf("decode trailing YAML: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return pkgBenchmark.Manifest{}, err
	}

	return manifest, nil
}
