//go:build darwin || freebsd || linux || netbsd || windows

package opuspurego

import (
	"fmt"
	"runtime"

	"github.com/ebitengine/purego"
)

type nativeFunctions struct {
	version        func() string
	encoderCreate  func(int32, int32, int32, *int32) uintptr
	encoderDestroy func(uintptr)
	decoderCreate  func(int32, int32, *int32) uintptr
	decoderDestroy func(uintptr)
	encode         func(uintptr, *int16, int32, *byte, int32) int32
	encodeFloat    func(uintptr, *float32, int32, *byte, int32) int32
	decode         func(uintptr, *byte, int32, *int16, int32, int32) int32
	decodeFloat    func(uintptr, *byte, int32, *float32, int32, int32) int32
	packetSamples  func(*byte, int32, int32) int32
	encoderSet     func(uintptr, int32, int32) int32
	encoderGet     func(uintptr, int32, *int32) int32
	decoderSet     func(uintptr, int32, int32) int32
}

func bindNativeFunctions(h uintptr) (f nativeFunctions, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("register native function: %v", p)
		}
	}()
	for _, b := range []struct {
		name string
		dst  any
	}{
		{"opus_get_version_string", &f.version}, {"opus_encoder_create", &f.encoderCreate},
		{"opus_encoder_destroy", &f.encoderDestroy}, {"opus_decoder_create", &f.decoderCreate},
		{"opus_decoder_destroy", &f.decoderDestroy}, {"opus_encode", &f.encode},
		{"opus_encode_float", &f.encodeFloat}, {"opus_decode", &f.decode},
		{"opus_decode_float", &f.decodeFloat}, {"opus_packet_get_nb_samples", &f.packetSamples},
	} {
		p, e := lookupLibrarySymbol(h, b.name)
		if e != nil {
			return f, fmt.Errorf("resolve %s: %w", b.name, e)
		}
		purego.RegisterFunc(b.dst, p)
	}
	// Apple's ARM64 ABI puts C variadic arguments on the stack. Explicit
	// padding fills x2-x7 so purego places the control argument there.
	bindSet := func(name string) (func(uintptr, int32, int32) int32, error) {
		p, e := lookupLibrarySymbol(h, name)
		if e != nil {
			return nil, e
		}
		if (runtime.GOOS == "darwin" || runtime.GOOS == "ios") && runtime.GOARCH == "arm64" {
			var call func(uintptr, int32, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, int32) int32
			purego.RegisterFunc(&call, p)
			return func(s uintptr, r, v int32) int32 { return call(s, r, 0, 0, 0, 0, 0, 0, v) }, nil
		}
		var call func(uintptr, int32, int32) int32
		purego.RegisterFunc(&call, p)
		return func(s uintptr, r, v int32) int32 { return call(s, r, v) }, nil
	}
	if f.encoderSet, err = bindSet("opus_encoder_ctl"); err != nil {
		return
	}
	if f.decoderSet, err = bindSet("opus_decoder_ctl"); err != nil {
		return
	}
	p, e := lookupLibrarySymbol(h, "opus_encoder_ctl")
	if e != nil {
		return f, e
	}
	if (runtime.GOOS == "darwin" || runtime.GOOS == "ios") && runtime.GOARCH == "arm64" {
		var call func(uintptr, int32, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, *int32) int32
		purego.RegisterFunc(&call, p)
		f.encoderGet = func(s uintptr, r int32, v *int32) int32 { return call(s, r, 0, 0, 0, 0, 0, 0, v) }
	} else {
		var call func(uintptr, int32, *int32) int32
		purego.RegisterFunc(&call, p)
		f.encoderGet = func(s uintptr, r int32, v *int32) int32 { return call(s, r, v) }
	}
	return
}
