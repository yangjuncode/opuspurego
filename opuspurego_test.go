//go:build darwin || freebsd || linux || netbsd || windows

package opuspurego

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
)

func testLibrary(t *testing.T) *Library {
	t.Helper()
	var l *Library
	var err error
	if path := os.Getenv("LIBOPUS_PATH"); path != "" {
		l, err = Open(path)
	} else if root := os.Getenv("LIBOPUS_DIST"); root != "" {
		l, err = OpenDist(root)
	} else {
		l, err = OpenDefault()
		if err != nil {
			t.Skipf("libopus unavailable (set LIBOPUS_PATH or LIBOPUS_DIST to require it): %v", err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Error(err)
		}
	})
	return l
}
func TestRoundTrip(t *testing.T) {
	l := testLibrary(t)
	if version, err := l.Version(); err != nil || version == "" {
		t.Fatalf("version=%q error=%v", version, err)
	}
	for _, rate := range []int{8000, 12000, 16000, 24000, 48000} {
		for _, channels := range []int{1, 2} {
			t.Run(fmt.Sprintf("%d/%d", rate, channels), func(t *testing.T) {
				enc, err := l.NewEncoder(rate, channels, ApplicationAudio)
				if err != nil {
					t.Fatal(err)
				}
				defer enc.Close()
				dec, err := l.NewDecoder(rate, channels)
				if err != nil {
					t.Fatal(err)
				}
				defer dec.Close()
				for _, err := range []error{enc.SetBitrate(32000 * channels), enc.SetComplexity(5), enc.SetVBR(true), enc.SetDTX(false), enc.SetInBandFEC(true), enc.SetPacketLossPercent(10)} {
					if err != nil {
						t.Fatal(err)
					}
				}
				if got, err := enc.Bitrate(); err != nil || got != 32000*channels {
					t.Fatalf("bitrate=%d %v", got, err)
				}
				if got, err := enc.Complexity(); err != nil || got != 5 {
					t.Fatalf("complexity=%d %v", got, err)
				}
				if got, err := enc.Lookahead(); err != nil || got <= 0 {
					t.Fatalf("lookahead=%d %v", got, err)
				}
				frame := rate / 50
				pcm := make([]int16, frame*channels)
				dst := make([]int16, len(pcm))
				packet := make([]byte, 4000)
				var energy float64
				for k := 0; k < 5; k++ {
					for i := range pcm {
						pcm[i] = int16(12000 * math.Sin(2*math.Pi*440*float64(k*frame+i/channels)/float64(rate)))
					}
					n, err := enc.Encode(packet, pcm)
					if err != nil || n <= 0 {
						t.Fatalf("encode=%d %v", n, err)
					}
					if got, err := l.PacketSamples(packet[:n], rate); err != nil || got != frame {
						t.Fatalf("duration=%d %v", got, err)
					}
					got, err := dec.Decode(dst, packet[:n], false)
					if err != nil || got != frame {
						t.Fatalf("decode=%d %v", got, err)
					}
					for _, v := range dst {
						energy += float64(v) * float64(v)
					}
					if k == 4 {
						if got, err := dec.Decode(dst, packet[:n], true); err != nil || got != frame {
							t.Fatalf("FEC=%d %v", got, err)
						}
					}
				}
				if energy == 0 {
					t.Fatal("decoded silence")
				}
				if n, err := dec.DecodeLost(dst); err != nil || n != frame {
					t.Fatalf("PLC=%d %v", n, err)
				}
				fpcm := make([]float32, len(pcm))
				for i, v := range pcm {
					fpcm[i] = float32(v) / 32768
				}
				n, err := enc.EncodeFloat(packet, fpcm)
				if err != nil {
					t.Fatal(err)
				}
				if got, err := dec.DecodeFloat(fpcm, packet[:n], false); err != nil || got != frame {
					t.Fatalf("float decode=%d %v", got, err)
				}
				if err := enc.Reset(); err != nil {
					t.Fatal(err)
				}
				if err := dec.Reset(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
func TestLifetimeAndValidation(t *testing.T) {
	l := testLibrary(t)
	if _, err := l.NewEncoder(44100, 1, ApplicationAudio); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	enc, err := l.NewEncoder(48000, 2, ApplicationVoIP)
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	dec, err := l.NewDecoder(48000, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	if err := l.Close(); !errors.Is(err, ErrLibraryInUse) {
		t.Fatal(err)
	}
	for _, pcm := range [][]int16{nil, make([]int16, 1), make([]int16, 961), make([]int16, 960*2+2)} {
		if _, err := enc.Encode(make([]byte, 1000), pcm); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("len=%d error=%v", len(pcm), err)
		}
	}
	if _, err := enc.Encode(nil, make([]int16, 1920)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := dec.Decode(nil, nil, false); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := dec.Decode(make([]int16, 3), nil, false); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if err := enc.SetComplexity(11); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := l.PacketSamples(nil, 48000); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	packet := make([]byte, 4000)
	n, err := enc.Encode(packet, make([]int16, 1920))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dec.Decode(make([]int16, 2), packet[:n], false); err == nil {
		t.Fatal("small output accepted")
	}
	enc.Close()
	enc.Close()
	dec.Close()
	dec.Close()
	if _, err := enc.Encode(packet, make([]int16, 1920)); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := dec.DecodeLost(make([]int16, 1920)); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Version(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
	if _, err := l.NewDecoder(48000, 1); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
func TestConcurrentClose(t *testing.T) {
	l := testLibrary(t)
	enc, err := l.NewEncoder(48000, 1, ApplicationVoIP)
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, err := enc.Encode(make([]byte, 4000), make([]int16, 960))
				if err != nil && !errors.Is(err, ErrClosed) {
					t.Error(err)
				}
				err = l.Close()
				if err != nil && !errors.Is(err, ErrLibraryInUse) {
					t.Error(err)
				}
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); enc.Close() }()
	wg.Wait()
}
func TestOpenValidation(t *testing.T) {
	if _, err := Open(""); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := Open("bad\x00path"); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	if _, err := Open("/nonexistent/libopus.so"); err == nil {
		t.Fatal("missing library accepted")
	}
	if _, err := DistLibraryPath(""); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
}
