package retryamplification

func Call(layers, attempts int, leaf func() error) error {
	if layers == 0 {
		return leaf()
	}
	var err error
	for range attempts {
		err = Call(layers-1, attempts, leaf)
		if err == nil {
			return nil
		}
	}
	return err
}
