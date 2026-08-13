// Package buildinfo exposes release metadata without coupling product commands
// to the Benchmark Layer.
package buildinfo

import "runtime/debug"

var (
	Version = ""
	Commit  = ""
)

type Info struct {
	Version string
	Commit  string
	Dirty   bool
}

func Current() Info {
	current := Info{Version: Version, Commit: Commit}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return normalize(current)
	}
	if current.Version == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		current.Version = info.Main.Version
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if current.Commit == "" {
				current.Commit = setting.Value
			}
		case "vcs.modified":
			current.Dirty = setting.Value == "true"
		}
	}
	return normalize(current)
}

func normalize(info Info) Info {
	if info.Version == "" {
		info.Version = "devel"
	}
	if info.Commit == "" {
		info.Commit = "unknown"
	}
	return info
}
