package clierror

import "fmt"

// Wrap allows to cover and error with additional information
func Wrap(inside error, outside Error) Error {
	return WrapE(new(fmt.Sprintf("%v", inside)), outside)
}

// WrapE allows to cover clierror with additional information
func WrapE(inside, outside Error) Error {
	return inside.(*clierror).wrap(outside.(*clierror))
}
