package errs

import (
	"fmt"
	"path/filepath"
	"runtime"
)

type Error struct {
	Op   string
	File string
	Line int
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s:%d)", e.Op, e.File, e.Line)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	return &Error{
		Op:   op,
		File: filepath.Base(file),
		Line: line,
		Err:  err,
	}
}
