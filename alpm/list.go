package alpm

/*
#cgo LDFLAGS: -lalpm
#include <alpm.h>
*/
import "C"
import "unsafe"

// alpm_list_t
type List struct {
	data unsafe.Pointer
	prev *List
	next *List
}

func (list *List) ForEach(callback func(unsafe.Pointer)) {
	node := list
	for node = list ; node != nil ; node = node.next {
		callback(node.data)
	}
}
