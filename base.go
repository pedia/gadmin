package gadmin

func must[T any](xs ...any) T {
	// try return with (x, error)
	err, ok := xs[len(xs)-1].(error)
	if ok && err != nil {
		panic(err)
	}

	if !ok {
		// try return with (x, bool)
		if b, ok := xs[len(xs)-1].(bool); ok && !b {
			panic("not ok")
		}
	}

	return xs[0].(T)
}
