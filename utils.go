// Copyright (C) 2013 Space Monkey, Inc.

package errors

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"code.spacemonkey.com/go/space/log"
)

var (
	stackLogSize = flag.Int("errors.stack_trace_log_length", 4096,
		"The max stack trace byte length to log")

	logger = log.GetLoggerNamed("errors")
)

func LogWithStack(messages ...interface{}) {
	buf := make([]byte, *stackLogSize)
	buf = buf[:runtime.Stack(buf, false)]
	logger.Errorf("%s\n%s", fmt.Sprintln(messages...), buf)
}

func CatchPanic(err_ref *error) {
	r := recover()
	if r == nil {
		return
	}
	err, ok := r.(error)
	if ok {
		*err_ref = PanicError.Wrap(err)
		return
	}
	*err_ref = PanicError.New("%v", r)
}

type ErrorGroup struct {
	Errors []error
	limit  int
	excess int
}

func NewErrorGroup() *ErrorGroup { return &ErrorGroup{} }

func NewBoundedErrorGroup(limit int) *ErrorGroup {
	return &ErrorGroup{
		limit: limit,
	}
}

func (e *ErrorGroup) Add(err error) {
	if err != nil {
		if e.limit > 0 && len(e.Errors) == e.limit {
			e.excess++
		} else {
			e.Errors = append(e.Errors, err)
		}
	}
}

func (e *ErrorGroup) Finalize() error {
	if len(e.Errors) == 0 {
		return nil
	}
	if len(e.Errors) == 1 && e.excess == 0 {
		return e.Errors[0]
	}
	msgs := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	if e.excess > 0 {
		msgs = append(msgs, fmt.Sprintf("... and %d more.", e.excess))
		e.excess = 0
	}
	e.Errors = nil
	return ErrorGroupError.New(strings.Join(msgs, "\n"))
}

type LoggingErrorGroup struct {
	name   string
	total  int
	failed int
}

func NewLoggingErrorGroup(name string) *LoggingErrorGroup {
	return &LoggingErrorGroup{name: name}
}

func (e *LoggingErrorGroup) Add(err error) {
	e.total++
	if err != nil {
		logger.Errorf("%s: %s", e.name, err)
		e.failed++
	}
}

func (e *LoggingErrorGroup) Finalize() (err error) {
	if e.failed > 0 {
		err = ErrorGroupError.New("%s: %d of %d failed.", e.name, e.failed,
			e.total)
	}
	e.total = 0
	e.failed = 0
	return err
}

type Finalizer interface {
	Finalize() error
}

func Finalize(finalizers ...Finalizer) error {
	var errs ErrorGroup
	for _, finalizer := range finalizers {
		errs.Add(finalizer.Finalize())
	}
	return errs.Finalize()
}
