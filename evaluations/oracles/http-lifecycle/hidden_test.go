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

func TestDecodeBoundsContextAndBodyLifetime(t *testing.T) {
	tooLarge := &trackedBody{Reader: bytes.NewReader([]byte("123456"))}
	if _, err := Decode(context.Background(), tooLarge, 5); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("large body error = %v, want ErrTooLarge", err)
	}
	if !tooLarge.closed {
		t.Fatal("large body was not closed")
	}

	exact := &trackedBody{Reader: bytes.NewReader([]byte("12345"))}
	got, err := Decode(context.Background(), exact, 5)
	if err != nil || string(got) != "12345" || !exact.closed {
		t.Fatalf("exact body = %q, err=%v, closed=%v", got, err, exact.closed)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled := &trackedBody{Reader: bytes.NewReader([]byte("ok"))}
	if _, err := Decode(ctx, cancelled, 5); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled body error = %v", err)
	}
	if !cancelled.closed {
		t.Fatal("cancelled body was not closed")
	}

}
