//go:build windows

package tool

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const atomicTempPrefix = ".maestro-patch-"

const windowsFileRenameInformationEx = 65

func defaultAtomicFileOps() atomicFileOps { return windowsAtomicFileOps{} }

type windowsAtomicFileOps struct{ platformAtomicFileOps }

type windowsFileRenameInformation struct {
	Flags          uint32
	RootDirectory  windows.Handle
	FileNameLength uint32
	FileName       [1]uint16
}

type windowsFileDispositionInformation struct {
	DeleteFile byte
}

type windowsFileDispositionInformationEx struct {
	Flags uint32
}

type windowsFileBasicInformation struct {
	CreationTime   int64
	LastAccessTime int64
	LastWriteTime  int64
	ChangeTime     int64
	FileAttributes uint32
}

func (windowsAtomicFileOps) openTarget(parent *os.File, name string) (*os.File, error) {
	handle, err := openWindowsFileRelative(
		parent,
		name,
		windows.FILE_GENERIC_READ,
		windows.FILE_OPEN,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT,
	)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(handle), name)
	if file == nil {
		_ = windows.CloseHandle(handle)
		return nil, errors.New("construct target file")
	}
	return file, nil
}

func (windowsAtomicFileOps) createTemp(parent *os.File, mode os.FileMode) (*os.File, string, error) {
	for range 32 {
		var random [12]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, "", err
		}
		name := atomicTempPrefix + hex.EncodeToString(random[:])
		handle, err := openWindowsFileRelative(
			parent,
			name,
			windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE,
			windows.FILE_CREATE,
			windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_NON_DIRECTORY_FILE|
				windows.FILE_OPEN_REPARSE_POINT|windows.FILE_WRITE_THROUGH,
		)
		if errors.Is(err, syscall.ERROR_FILE_EXISTS) || errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
			continue
		}
		if err != nil {
			return nil, "", err
		}
		file := os.NewFile(uintptr(handle), name)
		if file == nil {
			_ = windows.CloseHandle(handle)
			_ = removeWindowsFileRelative(parent, name)
			return nil, "", errors.New("construct temporary file")
		}
		_ = mode // os.File.Chmod applies Windows' read-only permission bit later.
		return file, name, nil
	}
	return nil, "", errors.New("allocate unique patch temporary")
}

func (windowsAtomicFileOps) chmod(file *os.File, mode os.FileMode) error {
	var current windows.ByHandleFileInformation
	handle := windows.Handle(file.Fd())
	if err := windows.GetFileInformationByHandle(handle, &current); err != nil {
		return fmt.Errorf("read temporary attributes: %w", err)
	}
	information := windowsFileBasicInformation{FileAttributes: current.FileAttributes}
	if mode.Perm()&0o222 == 0 {
		information.FileAttributes |= windows.FILE_ATTRIBUTE_READONLY
	} else {
		information.FileAttributes &^= windows.FILE_ATTRIBUTE_READONLY
	}
	if information.FileAttributes == 0 {
		information.FileAttributes = windows.FILE_ATTRIBUTE_NORMAL
	}
	if err := windows.SetFileInformationByHandle(
		handle,
		windows.FileBasicInfo,
		(*byte)(unsafe.Pointer(&information)),
		uint32(unsafe.Sizeof(information)),
	); err != nil {
		return fmt.Errorf("set temporary attributes: %w", err)
	}
	return nil
}

func (windowsAtomicFileOps) rename(parent *os.File, source, target string) (bool, error) {
	handle, err := openWindowsFileRelative(
		parent,
		source,
		windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE,
		windows.FILE_OPEN,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT,
	)
	if err != nil {
		return false, fmt.Errorf("open atomic temporary for rename: %w", err)
	}

	targetUTF16, err := windows.UTF16FromString(target)
	if err != nil {
		_ = windows.CloseHandle(handle)
		return false, err
	}
	nameLength := (len(targetUTF16) - 1) * 2
	var layout windowsFileRenameInformation
	buffer := make([]byte, int(unsafe.Offsetof(layout.FileName))+nameLength)
	info := (*windowsFileRenameInformation)(unsafe.Pointer(&buffer[0]))
	info.Flags = windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS |
		windows.FILE_RENAME_IGNORE_READONLY_ATTRIBUTE
	info.RootDirectory = windows.Handle(parent.Fd())
	info.FileNameLength = uint32(nameLength)
	copy(unsafe.Slice(&info.FileName[0], len(targetUTF16)-1), targetUTF16[:len(targetUTF16)-1])

	err = windows.NtSetInformationFile(
		handle,
		&windows.IO_STATUS_BLOCK{},
		&buffer[0],
		uint32(len(buffer)),
		windowsFileRenameInformationEx,
	)
	if err != nil {
		_ = windows.CloseHandle(handle)
		return false, fmt.Errorf("rename atomic temporary: %w", windowsFilesystemError(err))
	}

	// This handle continues to identify the replacement after the namespace
	// operation. A later failure is therefore reported as committed, not unchanged.
	flushErr := windows.FlushFileBuffers(handle)
	closeErr := windows.CloseHandle(handle)
	if flushErr != nil || closeErr != nil {
		return true, fmt.Errorf("flush renamed target: %w", errors.Join(flushErr, closeErr))
	}
	return true, nil
}

func (windowsAtomicFileOps) syncDirectory(*os.File) error {
	// Windows has no Unix directory-fsync equivalent for this handle. The
	// replacement itself is opened write-through and flushed across the rename.
	return nil
}

func (windowsAtomicFileOps) remove(parent *os.File, name string) error {
	return removeWindowsFileRelative(parent, name)
}

func openWindowsFileRelative(
	parent *os.File,
	name string,
	access, disposition, options uint32,
) (windows.Handle, error) {
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return windows.InvalidHandle, err
	}
	attributes := &windows.OBJECT_ATTRIBUTES{
		RootDirectory: windows.Handle(parent.Fd()),
		ObjectName:    objectName,
		Attributes:    windows.OBJ_DONT_REPARSE,
	}
	attributes.Length = uint32(unsafe.Sizeof(*attributes))
	var handle windows.Handle
	err = windows.NtCreateFile(
		&handle,
		access|windows.SYNCHRONIZE,
		attributes,
		&windows.IO_STATUS_BLOCK{},
		nil,
		windows.FILE_ATTRIBUTE_NORMAL,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		disposition,
		options|windows.FILE_OPEN_FOR_BACKUP_INTENT,
		0,
		0,
	)
	if err != nil {
		return windows.InvalidHandle, windowsFilesystemError(err)
	}
	return handle, nil
}

func removeWindowsFileRelative(parent *os.File, name string) error {
	handle, err := openWindowsFileRelative(
		parent,
		name,
		windows.DELETE,
		windows.FILE_OPEN,
		windows.FILE_SYNCHRONOUS_IO_NONALERT|windows.FILE_OPEN_REPARSE_POINT,
	)
	if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
		return nil
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	extended := windowsFileDispositionInformationEx{Flags: windows.FILE_DISPOSITION_DELETE |
		windows.FILE_DISPOSITION_FORCE_IMAGE_SECTION_CHECK |
		windows.FILE_DISPOSITION_POSIX_SEMANTICS |
		windows.FILE_DISPOSITION_IGNORE_READONLY_ATTRIBUTE}
	err = windows.NtSetInformationFile(
		handle,
		&windows.IO_STATUS_BLOCK{},
		(*byte)(unsafe.Pointer(&extended)),
		uint32(unsafe.Sizeof(extended)),
		windows.FileDispositionInformationEx,
	)
	if err == nil {
		return nil
	}

	legacy := windowsFileDispositionInformation{DeleteFile: 1}
	err = windows.NtSetInformationFile(
		handle,
		&windows.IO_STATUS_BLOCK{},
		(*byte)(unsafe.Pointer(&legacy)),
		uint32(unsafe.Sizeof(legacy)),
		windows.FileDispositionInformation,
	)
	return windowsFilesystemError(err)
}

func windowsFilesystemError(err error) error {
	if status, ok := err.(windows.NTStatus); ok {
		return status.Errno()
	}
	return err
}
