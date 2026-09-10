//go:build darwin || freebsd || linux || netbsd || windows

package opuspurego

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Library owns a loaded libopus. Do not copy a Library after first use.
type Library struct {
	mu     sync.Mutex
	path   string
	handle uintptr
	refs   int
	fn     nativeFunctions
}

func Open(path string) (*Library, error) {
	if path == "" || strings.ContainsRune(path, 0) {
		return nil, ErrInvalidArgument
	}
	h, e := openLibrary(path)
	if e != nil {
		return nil, fmt.Errorf("opuspurego: open %q: %w", path, e)
	}
	f, e := bindNativeFunctions(h)
	if e != nil {
		_ = closeLibrary(h)
		return nil, fmt.Errorf("opuspurego: bind %q: %w", path, e)
	}
	return &Library{path: path, handle: h, fn: f}, nil
}
func OpenDefault() (*Library, error) { return Open(defaultLibraryName()) }
func DistTarget() (string, error) {
	prefix := "linux-"
	switch runtime.GOOS {
	case "linux", "android":
	case "windows":
		prefix = "mingw-"
	default:
		return "", ErrUnsupportedPlatform
	}
	arch := map[string]string{"amd64": "x64", "386": "x86", "arm64": "arm64", "arm": "armv7a"}[runtime.GOARCH]
	if arch == "" || (prefix == "mingw-" && arch != "x64" && arch != "x86") {
		return "", ErrUnsupportedPlatform
	}
	return prefix + arch, nil
}
func DistLibraryPath(root string) (string, error) {
	if root == "" {
		return "", ErrInvalidArgument
	}
	t, e := DistTarget()
	if e != nil {
		return "", e
	}
	name := "libopus.so"
	if runtime.GOOS == "windows" {
		name = "libopus.dll"
	}
	return filepath.Join(root, t, name), nil
}
func OpenDist(root string) (*Library, error) {
	p, e := DistLibraryPath(root)
	if e != nil {
		return nil, e
	}
	return Open(p)
}
func (l *Library) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}
func (l *Library) Version() (string, error) {
	if l == nil {
		return "", ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 {
		return "", ErrClosed
	}
	return l.fn.version(), nil
}
func (l *Library) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 {
		return nil
	}
	if l.refs != 0 {
		return fmt.Errorf("%w: %d codec(s)", ErrLibraryInUse, l.refs)
	}
	if e := closeLibrary(l.handle); e != nil {
		return e
	}
	l.handle = 0
	return nil
}
func (l *Library) NewEncoder(rate, channels int, application Application) (*Encoder, error) {
	if !validConfig(rate, channels) || (application != ApplicationVoIP && application != ApplicationAudio && application != ApplicationRestrictedLowDelay) {
		return nil, ErrInvalidArgument
	}
	if l == nil {
		return nil, ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 {
		return nil, ErrClosed
	}
	var status int32
	p := l.fn.encoderCreate(int32(rate), int32(channels), int32(application), &status)
	if _, e := nativeStatus("encoder_create", status); e != nil {
		return nil, e
	}
	if p == 0 {
		return nil, &NativeError{"encoder_create", -7}
	}
	l.refs++
	e := &Encoder{library: l, ptr: p, rate: rate, channels: channels}
	runtime.SetFinalizer(e, (*Encoder).finalize)
	return e, nil
}
func (l *Library) NewDecoder(rate, channels int) (*Decoder, error) {
	if !validConfig(rate, channels) {
		return nil, ErrInvalidArgument
	}
	if l == nil {
		return nil, ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 {
		return nil, ErrClosed
	}
	var status int32
	p := l.fn.decoderCreate(int32(rate), int32(channels), &status)
	if _, e := nativeStatus("decoder_create", status); e != nil {
		return nil, e
	}
	if p == 0 {
		return nil, &NativeError{"decoder_create", -7}
	}
	l.refs++
	d := &Decoder{library: l, ptr: p, rate: rate, channels: channels}
	runtime.SetFinalizer(d, (*Decoder).finalize)
	return d, nil
}
func (l *Library) releaseCodec() { l.mu.Lock(); l.refs--; l.mu.Unlock() }

// PacketSamples returns the packet duration in samples per channel at rate.
func (l *Library) PacketSamples(packet []byte, rate int) (int, error) {
	if len(packet) == 0 || len(packet) > 2147483647 || !validConfig(rate, 1) {
		return 0, ErrInvalidArgument
	}
	if l == nil {
		return 0, ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 {
		return 0, ErrClosed
	}
	n := l.fn.packetSamples(&packet[0], int32(len(packet)), int32(rate))
	runtime.KeepAlive(packet)
	return nativeStatus("packet_get_nb_samples", n)
}
