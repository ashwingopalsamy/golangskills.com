package apicompatibility

import "testing"

func TestExistingConsumerStillCompiles(t *testing.T) {
	request := Request{UserID: "user-1"}
	if got := Subject(request); got != "user-1" {
		t.Fatalf("Subject() = %q", got)
	}
}
