package clierror

func Wrap(inside error, outside *Error) error {
	if err, ok := inside.(*Error); ok {
		return err.wrap(outside)
	} else {
		return &Error{
			Message: outside.Message,
			Details: wrapDetails(outside.Details, inside.Error()),
			Hints:   outside.Hints,
		}
	}
}
