package clierror

import "fmt"

type modifier func(err *clierror)

func Message(msg string) modifier {
	return MessageF("%s", msg)
}

func MessageF(format string, args ...interface{}) modifier {
	return func(err *clierror) {
		err.message = fmt.Sprintf(format, args...)
	}
}

func Hints(hints ...string) modifier {
	return func(err *clierror) {
		err.hints = hints
	}
}
