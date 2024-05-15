package clierror

import (
	"fmt"
	"os"
)

// Check prints error and executes os.Exit(1) if the error is not nil
func Check(err Error) {
	if err != nil {
		fmt.Println(err.String())
		os.Exit(1)
	}
}
