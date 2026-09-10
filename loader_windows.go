package opuspurego

import "syscall"

func openLibrary(path string) (uintptr, error) {
	h, e := syscall.LoadLibrary(path)
	return uintptr(h), e
}
func lookupLibrarySymbol(h uintptr, n string) (uintptr, error) {
	return syscall.GetProcAddress(syscall.Handle(h), n)
}
func closeLibrary(h uintptr) error { return syscall.FreeLibrary(syscall.Handle(h)) }
func defaultLibraryName() string   { return "libopus.dll" }
