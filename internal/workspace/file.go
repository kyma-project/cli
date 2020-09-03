package workspace

import "io"

type file interface {
	write(io.Writer, interface{}) error
	fileName() string
}
