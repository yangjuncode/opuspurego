//go:build darwin || freebsd || linux || netbsd || windows

package opuspurego

import (
	"runtime"
	"sync"
)

// Encoder owns a serialized audio stream. Do not copy it after first use.
type Encoder struct {
	mu             sync.Mutex
	library        *Library
	ptr            uintptr
	rate, channels int
}

func (e *Encoder) validFrame(samples int) bool {
	if samples%e.channels != 0 {
		return false
	}
	n := samples / e.channels
	for _, v := range []int{e.rate / 400, e.rate / 200, e.rate / 100, e.rate / 50, e.rate / 25, e.rate * 3 / 50} {
		if n == v {
			return true
		}
	}
	return false
}

// Encode encodes interleaved PCM. It returns bytes written to dst.
func (e *Encoder) Encode(dst []byte, pcm []int16) (int, error) {
	if e == nil {
		return 0, ErrClosed
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ptr == 0 {
		return 0, ErrClosed
	}
	if len(dst) == 0 || len(dst) > 2147483647 || !e.validFrame(len(pcm)) {
		return 0, ErrInvalidArgument
	}
	n := e.library.fn.encode(e.ptr, &pcm[0], int32(len(pcm)/e.channels), &dst[0], int32(len(dst)))
	runtime.KeepAlive(pcm)
	runtime.KeepAlive(dst)
	runtime.KeepAlive(e)
	return nativeStatus("encode", n)
}

// EncodeFloat accepts interleaved PCM nominally in [-1, 1].
func (e *Encoder) EncodeFloat(dst []byte, pcm []float32) (int, error) {
	if e == nil {
		return 0, ErrClosed
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ptr == 0 {
		return 0, ErrClosed
	}
	if len(dst) == 0 || len(dst) > 2147483647 || !e.validFrame(len(pcm)) {
		return 0, ErrInvalidArgument
	}
	n := e.library.fn.encodeFloat(e.ptr, &pcm[0], int32(len(pcm)/e.channels), &dst[0], int32(len(dst)))
	runtime.KeepAlive(pcm)
	runtime.KeepAlive(dst)
	runtime.KeepAlive(e)
	return nativeStatus("encode_float", n)
}
func (e *Encoder) set(request int32, value int) error {
	if e == nil {
		return ErrClosed
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ptr == 0 {
		return ErrClosed
	}
	if int64(value) < -2147483648 || int64(value) > 2147483647 {
		return ErrInvalidArgument
	}
	_, err := nativeStatus("encoder_ctl", e.library.fn.encoderSet(e.ptr, request, int32(value)))
	runtime.KeepAlive(e)
	return err
}
func (e *Encoder) get(request int32) (int, error) {
	if e == nil {
		return 0, ErrClosed
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ptr == 0 {
		return 0, ErrClosed
	}
	var value int32
	_, err := nativeStatus("encoder_ctl", e.library.fn.encoderGet(e.ptr, request, &value))
	runtime.KeepAlive(e)
	return int(value), err
}
func (e *Encoder) SetBitrate(bps int) error      { return e.set(4002, bps) }
func (e *Encoder) Bitrate() (int, error)         { return e.get(4003) }
func (e *Encoder) SetComplexity(value int) error { return e.set(4010, value) }
func (e *Encoder) Complexity() (int, error)      { return e.get(4011) }
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func (e *Encoder) SetVBR(v bool) error                  { return e.set(4006, boolInt(v)) }
func (e *Encoder) SetInBandFEC(v bool) error            { return e.set(4012, boolInt(v)) }
func (e *Encoder) SetPacketLossPercent(value int) error { return e.set(4014, value) }
func (e *Encoder) SetDTX(v bool) error                  { return e.set(4016, boolInt(v)) }
func (e *Encoder) Lookahead() (int, error)              { return e.get(4027) }
func (e *Encoder) Reset() error                         { return e.set(4028, 0) }
func (e *Encoder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.ptr == 0 {
		return nil
	}
	e.library.fn.encoderDestroy(e.ptr)
	e.ptr = 0
	e.library.releaseCodec()
	runtime.SetFinalizer(e, nil)
	return nil
}
func (e *Encoder) finalize() { _ = e.Close() }
