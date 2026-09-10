package opuspurego

import (
	"errors"
	"fmt"
)

var (
	ErrClosed              = errors.New("opuspurego: resource is closed")
	ErrLibraryInUse        = errors.New("opuspurego: library is still in use")
	ErrInvalidArgument     = errors.New("opuspurego: invalid argument")
	ErrUnsupportedPlatform = errors.New("opuspurego: unsupported build target")
)

// NativeError preserves the negative error code returned by libopus.
type NativeError struct {
	Operation string
	Code      int
}

func (e *NativeError) Error() string {
	return fmt.Sprintf("opuspurego: %s: opus status %d", e.Operation, e.Code)
}
func (e *NativeError) Unwrap() error {
	if e.Code == -1 {
		return ErrInvalidArgument
	}
	return nil
}
func nativeStatus(op string, n int32) (int, error) {
	if n < 0 {
		return int(n), &NativeError{op, int(n)}
	}
	return int(n), nil
}

type Application int32

const (
	ApplicationVoIP               Application = 2048
	ApplicationAudio              Application = 2049
	ApplicationRestrictedLowDelay Application = 2051
	BitrateAuto                               = -1000
	BitrateMax                                = -1
)

func validConfig(rate, channels int) bool {
	return (channels == 1 || channels == 2) && (rate == 8000 || rate == 12000 || rate == 16000 || rate == 24000 || rate == 48000)
}
