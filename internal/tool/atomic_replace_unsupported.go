//go:build !linux

package tool

import (
	"errors"
	"os"
)

var errAtomicReplaceUnsupported = errors.New("atomic workspace patch is supported only on linux")

func defaultAtomicFileOps() atomicFileOps { return unsupportedAtomicFileOps{} }

type unsupportedAtomicFileOps struct{ platformAtomicFileOps }

func (unsupportedAtomicFileOps) openTarget(*os.File, string) (*os.File, error) {
	return nil, errAtomicReplaceUnsupported
}
func (unsupportedAtomicFileOps) createTemp(*os.File, os.FileMode) (*os.File, string, error) {
	return nil, "", errAtomicReplaceUnsupported
}
func (unsupportedAtomicFileOps) rename(*os.File, string, string) error {
	return errAtomicReplaceUnsupported
}
func (unsupportedAtomicFileOps) remove(*os.File, string) error {
	return errAtomicReplaceUnsupported
}
