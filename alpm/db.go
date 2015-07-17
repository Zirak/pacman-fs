package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"unsafe"
)

func (hand Handle) GetLocalDb() (*DB, error) {
	db := C.alpm_get_localdb(hand.ptr)

	if db == nil {
		return nil, hand.Error()
	}

	return &DB{db, hand}, nil
}

func (hand Handle) RegisterSyncDb(dbname string) (*DB, error) {
	cdbname := C.CString(dbname)
	defer C.free(unsafe.Pointer(cdbname))

	// XXX siglevel argument
	db := C.alpm_register_syncdb(hand.ptr, cdbname, 0)

	if db == nil {
		return nil, hand.Error()
	}

	return &DB{db, hand}, nil
}

func (db DB) GetPkgcache() *PkgList {
	// This isn't pretty. get_pkgcache returns a pointer to a alpm_list.
	//we've defined that struct, but we need to convert that pointer to it.

	cache := (*List)(unsafe.Pointer(C.alpm_db_get_pkgcache(db.ptr)))
	// XXX list_to_slice ?
	return &PkgList{cache, db.handle}
}
