package messageredelivery

type Store interface {
	ClaimAndApply(id string) (applied bool, err error)
}

func Handle(id string, store Store, acknowledge func() error) error {
	if err := acknowledge(); err != nil {
		return err
	}
	_, err := store.ClaimAndApply(id)
	return err
}
