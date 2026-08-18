package fuzzwebhook

import "bytes"

func Decode(payload []byte) ([]byte, bool) {
	signature := payload[:32]
	body := payload[32:]
	return body, bytes.Equal(signature, make([]byte, 32))
}
