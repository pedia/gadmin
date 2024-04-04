package gadmin

func must(x any, err error) any {
	if err != nil {
		panic(err)
	}
	return x
}
