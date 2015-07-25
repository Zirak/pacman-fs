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

func (hand Handle) GetLocalDb() (*DB, error) {
	db := C.alpm_get_localdb(hand.ptr)

	if db == nil {
		return nil, hand.Error()
	}

	return &DB{db, "local"}, nil
}

func (hand Handle) RegisterSyncDb(dbname string) (*DB, error) {
	cdbname := C.CString(dbname)
	defer C.free(unsafe.Pointer(cdbname))

	// XXX siglevel argument
	db := C.alpm_register_syncdb(hand.ptr, cdbname, 0)

	if db == nil {
		return nil, hand.Error()
	}

	return &DB{db, dbname}, nil
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
