package ledgerbalance

import "errors"

type Entry struct {
	Currency string
	Amount   int64
}

func Validate(entries []Entry) error {
	var total int64
	for _, entry := range entries {
		total += entry.Amount
	}
	if total != 0 {
		return errors.New("unbalanced journal")
	}
	return nil
}
