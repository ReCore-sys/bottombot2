package logging

import (
	"errors"
	"os"

	"github.com/ztrue/tracerr"
)

// `Log` takes an error and prints it to the console with a stack trace
//
// @param error err The error you want to log
func Log(err error) {
	wrap := tracerr.Wrap(err)
	f, err := os.OpenFile("./static/log.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(tracerr.SprintSourceColor(wrap) + "\n")
	if err != nil {
		panic(err)
	}
	tracerr.PrintSourceColor(wrap)
}

// `LogString` takes a string and prints it to the console with a stack trace
//
// @param string err The error message to be logged.
func LogString(err string) {
	tracerr.PrintSourceColor(errors.New(err))
}
