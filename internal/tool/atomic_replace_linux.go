//go:build linux

package tool

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"syscall"
)

const atomicTempPrefix = ".maestro-patch-"

func defaultAtomicFileOps() atomicFileOps { return linuxAtomicFileOps{} }

type linuxAtomicFileOps struct{ platformAtomicFileOps }

func (linuxAtomicFileOps) openTarget(parent *os.File, name string) (*os.File, error) {
	fd, err := syscall.Openat(int(parent.Fd()), name, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, errors.New("construct target file")
	}
	return file, nil
}

func (linuxAtomicFileOps) createTemp(parent *os.File, mode os.FileMode) (*os.File, string, error) {
	for range 32 {
		var random [12]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, "", err
		}
		name := atomicTempPrefix + hex.EncodeToString(random[:])
		fd, err := syscall.Openat(
			int(parent.Fd()), name,
			syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_CLOEXEC|syscall.O_NOFOLLOW,
			uint32(mode.Perm()),
		)
		if errors.Is(err, syscall.EEXIST) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		file := os.NewFile(uintptr(fd), name)
		if file == nil {
			_ = syscall.Close(fd)
			_ = syscall.Unlinkat(int(parent.Fd()), name)
			return nil, "", errors.New("construct temporary file")
		}
		return file, name, nil
	}
	return nil, "", errors.New("allocate unique patch temporary")
}

func (linuxAtomicFileOps) rename(parent *os.File, source, target string) error {
	return syscall.Renameat(int(parent.Fd()), source, int(parent.Fd()), target)
}

func (linuxAtomicFileOps) remove(parent *os.File, name string) error {
	err := syscall.Unlinkat(int(parent.Fd()), name)
	if errors.Is(err, syscall.ENOENT) {
		return nil
	}
	return err
}
