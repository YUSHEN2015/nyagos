package dos

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var mpr = syscall.NewLazyDLL("mpr")
var wNetGetConnectionW = mpr.NewProc("WNetGetConnectionW")
var wNetOpenEnum = mpr.NewProc("WNetOpenEnumW")
var wNetEnumResource = mpr.NewProc("WNetEnumResourceW")
var wNetCloseEnum = mpr.NewProc("WNetCloseEnum")

func WNetGetConnection(localName string) (string, error) {
	localNamePtr, localNameErr := syscall.UTF16PtrFromString(localName)
	if localNameErr != nil {
		return "", localNameErr
	}
	var buffer [1024]uint16
	size := uintptr(len(buffer))

	rc, _, err := wNetGetConnectionW.Call(
		uintptr(unsafe.Pointer(localNamePtr)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)))

	if uint32(rc) != 0 {
		return "", err
	}
	return syscall.UTF16ToString(buffer[:]), nil
}

func NetDriveToUNC(path string) string {
	if path[1] == ':' {
		// print("'", path[:2], "'\n")
		_path, err := WNetGetConnection(path[:2])
		if err == nil {
			return filepath.Join(_path, path[2:])
		}
		// print(err.Error(), "\n")
	}
	return path
}

type netresourceT struct {
	Scope       uint32
	Type        uint32
	DisplayType uint32
	Usage       uint32
	LocalName   *uint16
	RemoteName  *uint16
	Comment     *uint16
	Provider    *uint16
	_           [16 * 1024]byte
}

func u2str(u *uint16) string {
	buffer := make([]uint16, 0, 100)
	for *u != 0 {
		buffer = append(buffer, *u)
		u = (*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(u)) + 1))
	}
	return syscall.UTF16ToString(buffer)
}

func WNetEnum(handler func(localName string, remoteName string)) error {
	var handle uintptr

	rc, _, err := wNetOpenEnum.Call(
		RESOURCE_GLOBALNET,
		RESOURCETYPE_DISK,
		RESOURCEUSAGE_CONTAINER,
		0,
		uintptr(unsafe.Pointer(&handle)))
	if rc != windows.NO_ERROR {
		return fmt.Errorf("NetOpenEnum: %s", err)
	}
	defer wNetCloseEnum.Call(handle)
	for {
		buffer := netresourceT{
			Scope:       RESOURCE_GLOBALNET,
			Type:        RESOURCETYPE_DISK,
			DisplayType: RESOURCEDISPLAYTYPE_NETWORK,
			Usage:       RESOURCEUSAGE_CONTAINER,
			LocalName:   nil,
			RemoteName:  nil,
			Comment:     nil,
			Provider:    nil,
		}
		size := unsafe.Sizeof(buffer)
		rc, _, err := wNetEnumResource.Call(
			handle,
			1,
			uintptr(unsafe.Pointer(&buffer)),
			uintptr(unsafe.Pointer(&size)))

		if rc == windows.NO_ERROR {
			handler(u2str(buffer.LocalName), u2str(buffer.RemoteName))
		} else if rc == ERROR_NO_MORE_ITEMS {
			return nil
		} else {
			return fmt.Errorf("NetEnumResource: %s", err)
		}
	}
}

// https://msdn.microsoft.com/ja-jp/library/cc447030.aspx
// http://eternalwindows.jp/security/share/share06.html
