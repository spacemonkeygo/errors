package errors

import (
    "fmt"
)

var root *ErrorClass

func New(ec *ErrorClass) *ErrorClass {
    return &ErrorClass{parent: ec}
}

func Wrap(err error) *Error {
    return root.Wrap(err)
}

type ErrorClass struct {
    parent *ErrorClass
}

func (e *ErrorClass) New(format string, args ...interface{}) *Error {
    return e.Wrap(fmt.Errorf(format, args...))
}

func (e *ErrorClass) Wrap(err error) *Error {
    return &Error{error: err, class: e}
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
    error
    class *ErrorClass
}

func (e *Error) Is(ec *ErrorClass) bool {
    return e.class.Is(ec)
}
