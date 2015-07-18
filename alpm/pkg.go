package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"unsafe"
	"log"
)

func pointerToDep(depptr *C.alpm_depend_t) PkgDep {
	return PkgDep{
		Name: C.GoString(depptr.name),
		Version: C.GoString(depptr.version),
		Desc: C.GoString(depptr.desc),
		NameHash: uint64(depptr.name_hash),
		Mod: int(depptr.mod),
	}
}

// XXX inline those functions into attributes on a Pkg
func (pkg Pkg) Name() string {
	return C.GoString(C.alpm_pkg_get_name(pkg.ptr))
}
func (pkg Pkg) Version() string {
	return C.GoString(C.alpm_pkg_get_version(pkg.ptr))
}
func (pkg Pkg) Desc() string {
	return C.GoString(C.alpm_pkg_get_desc(pkg.ptr))
}

func (pkg Pkg) InstallSize() int64 {
	return int64(C.alpm_pkg_get_isize(pkg.ptr))
}

// In alpm, a package has a Provides section, which is "a list of packages
//provided by the package". In other words, a package may have several names.
// This function returns whether the package provides a given package name.
func (pkg Pkg) Provides(name string) bool {
	if pkg.Name() == name {
		return true
	}

	for _, prov := range pkg.GetProvides() {
		if prov.Name == name {
			return true
		}
	}

	return false
}

func (pkg Pkg) GetDeps() []PkgDep {
	deps := []PkgDep{}

	uglyDeps := (*List)(unsafe.Pointer(C.alpm_pkg_get_depends(pkg.ptr)))

	uglyDeps.ForEach(func(depptr unsafe.Pointer) {
		dep := pointerToDep((*C.alpm_depend_t)(depptr))
		log.Println(dep)
		deps = append(deps, dep)
	})

	return deps
}

func (pkg Pkg) GetProvides() []PkgDep {
	provides := []PkgDep{}

	ugly := (*List)(unsafe.Pointer(C.alpm_pkg_get_provides(pkg.ptr)))

	ugly.ForEach(func(provptr unsafe.Pointer) {
		dep := pointerToDep((*C.alpm_depend_t)(provptr))
		log.Println("", dep)
		provides = append(provides, dep)
	})

	return provides
}

// XXX shit ton of other functions

func (pkglist PkgList) Slice() []Pkg {
	ret := []Pkg{}

	pkglist.ForEach(func(pkgptr unsafe.Pointer) {
		pkg := Pkg{(*C.alpm_pkg_t)(pkgptr), pkglist.handle}
		ret = append(ret, pkg)
	})

	return ret
}
