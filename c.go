package main

import "C"
import (
	"fmt"
	"syscall"
	"unsafe"
)

// goStrToCWideString(fmt.Sprintf("image (*%s)\x00*%s\x00All Files (*.*)\x00*.*\x00", ext, ext))
// 以上方法會遇到: EINVAL (536870951)的錯誤
// 因為中間給了\x00，因此要分開寫再補上\x00
func buildFilter(format string) ([]uint16, error) {
	var filterUTF16 []uint16
	for _, e := range []struct {
		Name    string
		Pattern string
	}{
		{"image", format},
		{"All Files", ".*"},
	} {

		displayName := fmt.Sprintf("%s (*%s)", e.Name, e.Pattern)
		displayNameUTF16, err := syscall.UTF16FromString(displayName) // 結尾自動會補上\x00也就是0 // image (*%s)\x00
		if err != nil {
			return nil, err
		}
		pattern := fmt.Sprintf("*%s", e.Pattern)
		patternUTF16, err := syscall.UTF16FromString(pattern) // *%s
		if err != nil {
			return nil, err
		}

		filterUTF16 = append(filterUTF16, displayNameUTF16...)
		filterUTF16 = append(filterUTF16, patternUTF16...)
	}
	// 結尾雙 NUL 结尾
	filterUTF16 = append(filterUTF16, 0)

	return filterUTF16, nil
}

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

func utf16ToCWideString(u16 []uint16) *C.wchar_t {
	size := len(u16) * int(unsafe.Sizeof(u16[0]))
	ptr := C.malloc(C.size_t(size))
	if ptr == nil {
		return nil
	}
	cptr := (*[1 << 30]C.wchar_t)(ptr)
	for i, v := range u16 {
		cptr[i] = C.wchar_t(v)
	}
	return (*C.wchar_t)(ptr)
}
