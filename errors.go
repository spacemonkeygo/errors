package errors

import (
    "fmt"
)

type ErrorClass struct {
    parent *ErrorClass
    name   string
}

var (
    // base error classes. To construct your own error class, use New.
    SystemError       = &ErrorClass{parent: nil, name: "System Error"}
    HierarchicalError = &ErrorClass{parent: nil, name: "Error"}
)

func New(ec *ErrorClass, name string) *ErrorClass {
    if ec == nil {
        ec = HierarchicalError
    }
    return &ErrorClass{parent: ec, name: name}
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

func (e *ErrorClass) Wrap(err error) error {
    if err == nil {
        return nil
    }
    return &Error{err: err, class: e}
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

func WrappedErr(err error) error {
    cast, ok := err.(*Error)
    if !ok {
        return nil
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
    NotImplementedError = New(nil, "Not Implemented Error")
)
