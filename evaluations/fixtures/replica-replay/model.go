package replicareplay

import "errors"

var ErrKeyConflict = errors.New("idempotency key conflicts with prior request")

type Request struct {
	Key     string
	Account string
	Amount  int64
}

type Result struct {
	EntryID int64
	Account string
	Amount  int64
}
