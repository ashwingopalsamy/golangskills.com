package commitambiguity

type Decision string

const (
	Retry     Decision = "retry"
	Committed Decision = "committed"
	Reconcile Decision = "reconcile"
)

func Decide(commitResponseLost, operationFound bool) Decision {
	if commitResponseLost {
		return Retry
	}
	if operationFound {
		return Committed
	}
	return Retry
}
