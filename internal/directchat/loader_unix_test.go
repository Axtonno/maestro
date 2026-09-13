//go:build unix

package directchat

import "syscall"

func createFileLoaderPipe(name string) (bool, error) {
	return true, syscall.Mkfifo(name, 0o600)
}
