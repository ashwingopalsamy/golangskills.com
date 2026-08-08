package retryamplification

import (
	"errors"
	"testing"
)

func TestOneLayerOwnsRetryBudget(t *testing.T) {
	tests := []struct {
		name     string
		layers   int
		attempts int
		failures int
		want     int
		wantErr  bool
	}{
		{name: "deep failure", layers: 5, attempts: 3, failures: 10, want: 3, wantErr: true},
		{name: "success stops retry", layers: 3, attempts: 5, failures: 1, want: 2},
		{name: "zero layers still uses owner budget", layers: 0, attempts: 3, failures: 10, want: 3, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			err := Call(test.layers, test.attempts, func() error {
				calls++
				if calls <= test.failures {
					return errors.New("down")
				}
				return nil
			})
			if calls != test.want {
				t.Fatalf("leaf attempts = %d, want %d", calls, test.want)
			}
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr=%v", err, test.wantErr)
			}
		})
	}
}
