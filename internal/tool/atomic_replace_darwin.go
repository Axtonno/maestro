//go:build darwin

package tool

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

const atomicTempPrefix = ".maestro-patch-"

func defaultAtomicFileOps() atomicFileOps { return darwinAtomicFileOps{} }

type darwinAtomicFileOps struct{ platformAtomicFileOps }

func (darwinAtomicFileOps) openTarget(parent *os.File, name string) (*os.File, error) {
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		_ = unix.Close(fd)
		return nil, errors.New("construct target file")
	}
	return file, nil
}

func (darwinAtomicFileOps) createTemp(parent *os.File, mode os.FileMode) (*os.File, string, error) {
	for range 32 {
		var random [12]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, "", err
		}
		name := atomicTempPrefix + hex.EncodeToString(random[:])
		fd, err := unix.Openat(
			int(parent.Fd()), name,
			unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW,
			uint32(mode.Perm()),
		)
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		file := os.NewFile(uintptr(fd), name)
		if file == nil {
			_ = unix.Close(fd)
			_ = unix.Unlinkat(int(parent.Fd()), name, 0)
			return nil, "", errors.New("construct temporary file")
		}
		return file, name, nil
	}
	return nil, "", errors.New("allocate unique patch temporary")
}

func (darwinAtomicFileOps) rename(parent *os.File, source, target string) error {
	return unix.Renameat(int(parent.Fd()), source, int(parent.Fd()), target)
}

func (darwinAtomicFileOps) remove(parent *os.File, name string) error {
	err := unix.Unlinkat(int(parent.Fd()), name, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	return err
}
