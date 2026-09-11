//go:build darwin || freebsd || linux || netbsd || windows

package opuspurego

import (
	"runtime"
	"sync"
)

// Decoder owns a serialized audio stream. Do not copy it after first use.
type Decoder struct {
	mu             sync.Mutex
	library        *Library
	ptr            uintptr
	rate, channels int
}

func (d *Decoder) validBuffer(samples, packetLen int, fec bool) bool {
	if samples == 0 || samples%d.channels != 0 || samples/d.channels > 2147483647 || packetLen > 2147483647 {
		return false
	}
	return (packetLen != 0 && !fec) || (samples/d.channels <= d.rate*120/1000 && (samples/d.channels)%(d.rate/400) == 0)
}

// Decode returns samples per channel. A nil/empty packet requests PLC using
// len(dst)/channels samples. For FEC pass the next packet and the missing
// frame's exact output length, then decode that packet normally as well.
func (d *Decoder) Decode(dst []int16, packet []byte, fec bool) (int, error) {
	if d == nil {
		return 0, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ptr == 0 {
		return 0, ErrClosed
	}
	if !d.validBuffer(len(dst), len(packet), fec) {
		return 0, ErrInvalidArgument
	}
	var p *byte
	if len(packet) > 0 {
		p = &packet[0]
	}
	n := d.library.fn.decode(d.ptr, p, int32(len(packet)), &dst[0], int32(len(dst)/d.channels), int32(boolInt(fec)))
	runtime.KeepAlive(packet)
	runtime.KeepAlive(dst)
	runtime.KeepAlive(d)
	return nativeStatus("decode", n)
}
func (d *Decoder) DecodeFloat(dst []float32, packet []byte, fec bool) (int, error) {
	if d == nil {
		return 0, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ptr == 0 {
		return 0, ErrClosed
	}
	if !d.validBuffer(len(dst), len(packet), fec) {
		return 0, ErrInvalidArgument
	}
	var p *byte
	if len(packet) > 0 {
		p = &packet[0]
	}
	n := d.library.fn.decodeFloat(d.ptr, p, int32(len(packet)), &dst[0], int32(len(dst)/d.channels), int32(boolInt(fec)))
	runtime.KeepAlive(packet)
	runtime.KeepAlive(dst)
	runtime.KeepAlive(d)
	return nativeStatus("decode_float", n)
}
func (d *Decoder) DecodeLost(dst []int16) (int, error) { return d.Decode(dst, nil, false) }
func (d *Decoder) Reset() error {
	if d == nil {
		return ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ptr == 0 {
		return ErrClosed
	}
	_, err := nativeStatus("decoder_ctl", d.library.fn.decoderSet(d.ptr, ctlResetState, 0))
	runtime.KeepAlive(d)
	return err
}
func (d *Decoder) Close() error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ptr == 0 {
		return nil
	}
	d.library.fn.decoderDestroy(d.ptr)
	d.ptr = 0
	d.library.releaseCodec()
	runtime.SetFinalizer(d, nil)
	return nil
}
func (d *Decoder) finalize() { _ = d.Close() }
