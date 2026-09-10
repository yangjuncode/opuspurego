//go:build darwin || freebsd || linux || netbsd

package opuspurego

import (
	"runtime"

	"github.com/ebitengine/purego"
)

func openLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
}
func lookupLibrarySymbol(h uintptr, n string) (uintptr, error) { return purego.Dlsym(h, n) }
func closeLibrary(h uintptr) error                             { return purego.Dlclose(h) }
func defaultLibraryName() string {
	if runtime.GOOS == "darwin" || runtime.GOOS == "ios" {
		return "libopus.dylib"
	}
	if runtime.GOOS == "android" {
		return "libopus.so"
	}
	return "libopus.so.0"
}
