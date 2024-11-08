package main

import "C"
import (
	"syscall"
	"unsafe"
)

// 解決: cgo argument has Go pointer to unpinned Go pointer
func goStrToCWideString(s string) *C.wchar_t {
	us, err := syscall.UTF16FromString(s)
	if err != nil {
		return nil
	}
	size := len(us) * int(unsafe.Sizeof(us[0]))
	ptr := C.malloc(C.size_t(size))
	if ptr == nil {
		return nil
	}
	cptr := (*[1 << 30]C.wchar_t)(ptr)
	for i, v := range us {
		cptr[i] = C.wchar_t(v)
	}
	return (*C.wchar_t)(ptr)
}

func wcharPtrToString(ptr *C.wchar_t) string {
	if ptr == nil {
		return ""
	}
	var u16s []uint16
	for i := 0; ; i++ {
		c := *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(i)*unsafe.Sizeof(*ptr)))
		if c == 0 {
			break
		}
		u16s = append(u16s, c)
	}
	return syscall.UTF16ToString(u16s)
}
