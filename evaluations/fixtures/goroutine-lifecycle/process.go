package goroutinelifecycle

import "context"

func ProcessAll(ctx context.Context, items []int, limit int, process func(context.Context, int) error) error {
	for _, item := range items {
		go process(ctx, item)
	}
	return nil
}
