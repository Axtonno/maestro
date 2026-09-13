package main

import (
	"errors"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func resolveExecutableForIdentity(executable string) (string, error) {
	resolved, err := filepath.EvalSymlinks(executable)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return "", err
	}
	pointer, pointerErr := windows.UTF16PtrFromString(executable)
	if pointerErr != nil {
		return "", pointerErr
	}
	attributes, attributeErr := windows.GetFileAttributes(pointer)
	if attributeErr != nil {
		return "", attributeErr
	}
	if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return "", err
	}
	return filepath.Clean(executable), nil
}
