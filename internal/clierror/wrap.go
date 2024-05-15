package clierror

func Wrap(inside any, modifiers ...modifier) Error {
	if err, ok := inside.(*clierror); ok {
		return err.wrap(new(modifiers...))
	}

	// convert any type of inside error to a string
	return new(MessageF("%v", inside)).wrap(new(modifiers...))
}
