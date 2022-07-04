package logging

import (
	"errors"

	"github.com/ztrue/tracerr"
)

// `Log` takes an error and prints it to the console with a stack trace
//
// @param error err The error you want to log
func Log(err error) {
	tracerr.PrintSourceColor(err)
}

// `LogString` takes a string and prints it to the console with a stack trace
//
// @param string err The error message to be logged.
func LogString(err string) {
	tracerr.PrintSourceColor(errors.New(err))
}
