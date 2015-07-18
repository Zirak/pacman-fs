package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type Handle struct {
	ptr *C.alpm_handle_t
}
type DB struct {
	ptr *C.alpm_db_t
	handle Handle
}
type Pkg struct {
	ptr *C.alpm_pkg_t
	handle Handle
}
type PkgDep struct {
	Name string
	Version string
	Desc string
	NameHash uint64
	Mod int
}
type PkgList struct {
	*List
	handle Handle
}

// alpm_list_t
type List struct {
	data unsafe.Pointer
	prev *List
	next *List
}

func Init() (*Handle, error) {
	c_root := C.CString("/")
	defer C.free(unsafe.Pointer(c_root))

	c_dbpath := C.CString("/var/lib/pacman")
	defer C.free(unsafe.Pointer(c_dbpath))

	var errno C.alpm_errno_t
	h := C.alpm_initialize(c_root, c_dbpath, &errno)

	if errno != 0 {
		return nil, strerror(errno)
	}

	return &Handle{h}, nil
}

func (hand Handle) Error() error {
	return strerror(C.alpm_errno(hand.ptr))
}

func (hand *Handle) Release() error {
	err := C.alpm_release(hand.ptr)

	if err != 0 {
		// XXX I didn't actually test whether alpm_errno actually returns a
		//relevant error when alpm_release fails...
		return hand.Error()
	}

	hand.ptr = nil
	return nil
}

func strerror(errno C.alpm_errno_t) error {
	return errors.New(C.GoString(C.alpm_strerror(errno)))
}

func (list *List) ForEach(callback func(unsafe.Pointer)) {
	node := list
	for node = list ; node != nil ; node = node.next {
		callback(node.data)
	}
}
