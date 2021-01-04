package cli

import "log"

// GetLogFunc returns the logger function used for CLI log output (used in Hydroform deployments)
func GetLogFunc(printLogs bool) func(format string, v ...interface{}) {
	if printLogs {
		return log.Printf
	}
	return func(format string, v ...interface{}) {
		//do nothing
	}
}
