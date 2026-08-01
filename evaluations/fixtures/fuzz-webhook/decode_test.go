package fuzzwebhook

import "testing"

func FuzzDecodeNeverPanics(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Fuzz(func(t *testing.T, payload []byte) {
		_, _ = Decode(payload)
	})
}
