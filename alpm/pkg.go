package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"unsafe"
)

func (pkg Pkg) Name() string {
	return C.GoString(C.alpm_pkg_get_name(pkg.ptr))
}
func (pkg Pkg) Version() string {
	return C.GoString(C.alpm_pkg_get_version(pkg.ptr))
}
func (pkg Pkg) Desc() string {
	return C.GoString(C.alpm_pkg_get_desc(pkg.ptr))
}

// XXX shit ton of other functions

func (pkglist PkgList) Slice() []Pkg {
	ret := []Pkg{}

	pkglist.ForEach(func (pkgptr unsafe.Pointer) {
		pkg := Pkg{(*C.alpm_pkg_t)(pkgptr), pkglist.handle}
		ret = append(ret, pkg)
	})

	return ret
}
