// Copyright (C) 2013 Space Monkey, Inc.

package errors

import (
    "fmt"

    "code.spacemonkey.com/go/space/log"
)

type ErrorClass struct {
    parent *ErrorClass
    name   string
    log    bool
}

var (
    // base error classes. To construct your own error class, use New.
    SystemError       = &ErrorClass{parent: nil, name: "System Error", log: false}
    HierarchicalError = &ErrorClass{parent: nil, name: "Error", log: false}
)

// NewWithLogging creates an error class for making specific errors.
// Additionally, whenever a specific error is generated, the
// current stack trace will be logged.
func NewWithLogging(ec *ErrorClass, name string) *ErrorClass {
    if ec == nil {
        ec = HierarchicalError
    }
    return &ErrorClass{parent: ec, name: name, log: true}
}

// NewWithoutLogging creates an error class for making specific errors.
// When errors from this class are generated, nothing is logged.
func NewWithoutLogging(ec *ErrorClass, name string) *ErrorClass {
    if ec == nil {
        ec = HierarchicalError
    }
    return &ErrorClass{parent: ec, name: name, log: false}
}

// New is like NewWithLogging or NewWithoutLogging, except the logging behavior
// is determined by the parent class. The two base classes of the error
// hierarchy (SystemError and HierarchicalError) do not log.
func New(ec *ErrorClass, name string) *ErrorClass {
    if ec == nil {
        ec = HierarchicalError
    }
    return &ErrorClass{parent: ec, name: name, log: ec.log}
}

func (e *ErrorClass) Is(parent *ErrorClass) bool {
    for check := e; check != nil; check = check.parent {
        if check == parent {
            return true
        }
    }
    return false
}

type Error struct {
    err   error
    class *ErrorClass
}

func (e *ErrorClass) Wrap(err error, classes ...*ErrorClass) error {
    if err == nil {
        return nil
    }
    ec, ok := err.(*Error)
    if !ok {
        rv := &Error{err: err, class: e}
        if e.log {
            log.PrintWithStack(rv.Error())
        }
        return rv
    }
    if ec.Is(e) {
        return err
    }
    for _, class := range classes {
        if ec.Is(class) {
            return err
        }
    }
    rv := &Error{err: err, class: e}
    if e.log {
        log.PrintWithStack(rv.Error())
    }
    return rv
}

func (e *ErrorClass) New(format string, args ...interface{}) error {
    return e.Wrap(fmt.Errorf(format, args...))
}

func (e *Error) Error() string {
    return fmt.Sprintf("%s: %s", e.class.name, e.err.Error())
}

func (e *Error) WrappedErr() error {
    return e.err
}

func (e *Error) Class() *ErrorClass {
    return e.class
}

func WrappedErr(err error) error {
    cast, ok := err.(*Error)
    if !ok {
        return err
    }
    return cast.WrappedErr()
}

func (e *Error) Is(ec *ErrorClass) bool {
    return e.class.Is(ec)
}

func (e *ErrorClass) Contains(err error) bool {
    cast, ok := err.(*Error)
    if !ok {
        return SystemError == e
    }
    return cast.Is(e)
}

var (
    // useful error classes
    NotImplementedError = NewWithLogging(nil, "Not Implemented Error")
    ProgrammerError     = NewWithLogging(nil, "Programmer Error")
)
