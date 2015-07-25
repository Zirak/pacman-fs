package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
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
