package reconciliationnetting

type Item struct {
	ID       string
	Currency string
	Amount   int64
}

type Group struct {
	Internal []Item
	External []Item
}

func Reconcile(internal, external []Item) []Group {
	var groups []Group
	for index := range internal {
		if index < len(external) && internal[index].Amount == external[index].Amount {
			groups = append(groups, Group{Internal: []Item{internal[index]}, External: []Item{external[index]}})
		}
	}
	return groups
}
