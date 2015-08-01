package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>

alpm_list_t *filelist_to_list(alpm_filelist_t *filelist) {
    alpm_list_t *ret = NULL;
    for (size_t i = 0; i < filelist->count; i += 1) {
        ret = alpm_list_add(ret, &filelist->files[i]);
    }
    return ret;
}
*/
import "C"
import "unsafe"

// alpm_list_t
type PointerList struct {
	data unsafe.Pointer
	prev *PointerList
	next *PointerList
}

func (list *PointerList) ForEach(callback func(unsafe.Pointer)) {
	node := list
	for node = list; node != nil; node = node.next {
		callback(node.data)
	}
}

type File struct {
	Name string
	Size uint64
	Mode uint64
}

func filelistToSlice(filelist *C.alpm_filelist_t) []*File {
	list := (*PointerList)(unsafe.Pointer(C.filelist_to_list(filelist)))
	files := []*File{}

	list.ForEach(func(fileptr unsafe.Pointer) {
		file := pointerToFile((*C.alpm_file_t)(fileptr))
		files = append(files, file)
	})

	return files
}

func pointerToFile(fileptr *C.alpm_file_t) *File {
	return &File{
		Name: C.GoString(fileptr.name),
		Size: uint64(fileptr.size),
		Mode: uint64(fileptr.mode),
	}
}
