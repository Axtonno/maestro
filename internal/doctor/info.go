package doctor

import (
	"os"
	"runtime"
)

type SystemInfo struct {
	OS      string
	Arch    string
	CPU     int
	HomeDir string
}

func CollectSystemInfo() SystemInfo {

	home, _ := os.UserHomeDir()

	return SystemInfo{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		CPU:     runtime.NumCPU(),
		HomeDir: home,
	}
}
