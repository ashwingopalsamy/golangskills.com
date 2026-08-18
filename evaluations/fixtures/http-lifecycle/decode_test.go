package httplifecycle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func TestDecodeBoundsAndClosesBody(t *testing.T) {
	body := &trackedBody{Reader: bytes.NewReader([]byte("123456"))}
	_, err := Decode(context.Background(), body, 5)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("error = %v, want ErrTooLarge", err)
	}
	if !body.closed {
		t.Fatal("body was not closed")
	}
}
