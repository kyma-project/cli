package clierror

func Wrap(inside error, modifiers ...modifier) error {
	if err, ok := inside.(*clierror); ok {
		return err.wrap(New(modifiers...))
	}

	return New(Message(inside.Error())).wrap(New(modifiers...))
}
