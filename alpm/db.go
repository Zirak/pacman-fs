package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"unsafe"
)

type DB struct {
	ptr *C.alpm_db_t
	Name string
}

func (db DB) FindPackage(name string) *Pkg {
	nameptr := C.CString(name)
	defer C.free(unsafe.Pointer(nameptr))

	pkgptr := C.alpm_db_get_pkg(db.ptr, nameptr)

	if pkgptr == nil {
		return nil
	}

	pkg := pointerToPkg(pkgptr)
	return &pkg
}

func (db DB) GetPkgcache() []Pkg {
	// This isn't pretty. get_pkgcache returns a pointer to a alpm_list.
	//we've defined that struct, but we need to convert that pointer to it.
	cache := (*PointerList)(unsafe.Pointer(C.alpm_db_get_pkgcache(db.ptr)))
	pkgs := []Pkg{}

	cache.ForEach(func(pkgptr unsafe.Pointer) {
		pkg := pointerToPkg((*C.alpm_pkg_t)(pkgptr))
		pkgs = append(pkgs, pkg)
	})

	return pkgs
}

func (db DB) GetProviderOf(name string) *Pkg {
	// XXX this is horribly inefficient - we go over the entire pkg cache for
	//every function call. Our caller, DepsDir, calls us *on every file lookup*
	// we need to generate a cache of some kind.
	pkgs := db.GetPkgcache()

	for _, pkg := range pkgs {
		if pkg.Provides(name) {
			return &pkg
		}
	}
	return nil
}
