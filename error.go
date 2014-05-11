package main

import (
	"fmt"
	"runtime"
)

type Error struct {
	message string
	cause   error
	tb      []byte
}

func NewError(message string, cause error) (err Error) {
	// 10K for the stacktrace
	err = Error{message, cause, make([]byte, 10240)}
	runtime.Stack(err.tb, false)
	return
}

func (e Error) Error() string {
	return e.message
}

func (e Error) FullFormat() string {
	result := fmt.Sprintf("error: %s\n", e.message)
	if e.cause != nil {
		result += fmt.Sprintf("caused by: %s\n", e.cause)
	}
	if PrintStacktrace {
		result += fmt.Sprintf("%s\n", string(e.tb))
	}
	return result
}
