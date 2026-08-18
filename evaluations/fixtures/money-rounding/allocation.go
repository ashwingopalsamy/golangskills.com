package moneyrounding

func Allocate(total int64, recipients []string) map[string]int64 {
	result := make(map[string]int64, len(recipients))
	for _, recipient := range recipients {
		result[recipient] = total / int64(len(recipients))
	}
	return result
}
