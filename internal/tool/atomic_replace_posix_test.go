//go:build linux || darwin

package tool

import "os"

func readAtomicTestFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
