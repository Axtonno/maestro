//go:build windows

package tool

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func readAtomicTestFile(name string) ([]byte, error) {
	nameUTF16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(
		nameUTF16,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), name)
	if file == nil {
		_ = windows.CloseHandle(handle)
		return nil, os.ErrInvalid
	}
	defer file.Close()
	return io.ReadAll(file)
}

func TestWindowsAtomicReplaceFailsClosedWhileTargetDeniesDeleteSharing(t *testing.T) {
	rootPath, logical, original := atomicFixture(t)
	target := filepath.Join(rootPath, filepath.FromSlash(logical))
	locked, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	proposed := "<?php\nfinal class Order {}\n"
	outcome, replaceErr := replacePhysicalFileAtomically(
		t.Context(), root, logical, digest(original), "class Order", "final class Order",
		proposed, 2<<20, defaultAtomicFileOps(),
	)
	_ = root.Close()
	if replaceErr == nil || outcome.committed {
		t.Fatalf("locked target did not fail closed: outcome=%#v err=%v", outcome, replaceErr)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != original {
		t.Fatalf("locked target changed: content=%q err=%v", content, err)
	}
	assertNoAtomicTemps(t, rootPath)
	if err := locked.Close(); err != nil {
		t.Fatal(err)
	}

	root, err = os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	outcome, replaceErr = replacePhysicalFileAtomically(
		t.Context(), root, logical, digest(original), "class Order", "final class Order",
		proposed, 2<<20, defaultAtomicFileOps(),
	)
	_ = root.Close()
	if replaceErr != nil || !outcome.committed || !outcome.durable {
		t.Fatalf("replacement after lock release failed: outcome=%#v err=%v", outcome, replaceErr)
	}
}
