package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"

import (
	"unsafe"
)

type Pkg struct {
	ptr *C.alpm_pkg_t

	Name        string
	Version     string
	Description string
	InstallSize int64
}

type PkgDep struct {
	Name        string
	Version     string
	Description string

	NameHash uint64
	Mod      int
}

// In alpm, a package has a Provides section, which is "a list of packages
//provided by the package". In other words, a package may have several names.
// This function returns whether the package provides a given package name.
func (pkg Pkg) Provides(name string) bool {
	if pkg.Name == name {
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

	uglyDeps := (*PointerList)(unsafe.Pointer(C.alpm_pkg_get_depends(pkg.ptr)))

	uglyDeps.ForEach(func(depptr unsafe.Pointer) {
		dep := pointerToDep((*C.alpm_depend_t)(depptr))
		deps = append(deps, dep)
	})

	return deps
}

func (pkg Pkg) GetProvides() []PkgDep {
	provides := []PkgDep{}

	ugly := (*PointerList)(unsafe.Pointer(C.alpm_pkg_get_provides(pkg.ptr)))

	ugly.ForEach(func(provptr unsafe.Pointer) {
		dep := pointerToDep((*C.alpm_depend_t)(provptr))
		provides = append(provides, dep)
	})

	return provides
}

// TODO return pointers, not structs

func pointerToPkg(pkgptr *C.alpm_pkg_t) Pkg {
	return Pkg{
		ptr:         pkgptr,

		Name:        C.GoString(C.alpm_pkg_get_name(pkgptr)),
		Version:     C.GoString(C.alpm_pkg_get_version(pkgptr)),
		Description: C.GoString(C.alpm_pkg_get_desc(pkgptr)),
		InstallSize: int64(C.alpm_pkg_get_isize(pkgptr)),
	}
}

func pointerToDep(depptr *C.alpm_depend_t) PkgDep {
	return PkgDep{
		Name:        C.GoString(depptr.name),
		Version:     C.GoString(depptr.version),
		Description: C.GoString(depptr.desc),

		NameHash:    uint64(depptr.name_hash),
		Mod:         int(depptr.mod),
	}
}
