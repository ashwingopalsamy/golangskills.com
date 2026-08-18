package httplifecycle

import (
	"context"
	"io"
)

var ErrTooLarge = io.ErrShortBuffer

func Decode(_ context.Context, body io.ReadCloser, _ int64) ([]byte, error) {
	return io.ReadAll(body)
}
