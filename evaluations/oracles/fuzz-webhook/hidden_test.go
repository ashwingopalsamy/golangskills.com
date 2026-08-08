package fuzzwebhook

import (
	"bytes"
	"testing"
)

func TestDecodeFramingBoundaries(t *testing.T) {
	for size := 0; size < 32; size++ {
		if body, ok := Decode(make([]byte, size)); ok || body != nil {
			t.Fatalf("size %d decoded as body=%x ok=%v", size, body, ok)
		}
	}
	payload := append(make([]byte, 32), []byte("body")...)
	body, ok := Decode(payload)
	if !ok || !bytes.Equal(body, []byte("body")) {
		t.Fatalf("valid frame decoded as body=%q ok=%v", body, ok)
	}
	payload[0] = 1
	body, ok = Decode(payload)
	if ok || !bytes.Equal(body, []byte("body")) {
		t.Fatalf("invalid signature changed framing: body=%q ok=%v", body, ok)
	}
}

func FuzzDecodeNeverPanics(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = Decode(payload)
	})
}
