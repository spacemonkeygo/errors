// Copyright (C) 2013 Space Monkey, Inc.

package errors

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var (
	logOnCreation      = GenSym()
	captureStack       = GenSym()
	disableInheritance = GenSym()
)

type ErrorClass struct {
	parent *ErrorClass
	name   string
	data   map[DataKey]interface{}
}

var (
	// base error classes. To construct your own error class, use New.
	SystemError = &ErrorClass{
		parent: nil,
		name:   "System Error",
		data:   map[DataKey]interface{}{}}
	HierarchicalError = &ErrorClass{
		parent: nil,
		name:   "Error",
		data:   map[DataKey]interface{}{captureStack: true}}
)

type ErrorOption func(map[DataKey]interface{})

func SetData(key DataKey, value interface{}) ErrorOption {
	return func(m map[DataKey]interface{}) {
		m[key] = value
	}
}

func LogOnCreation() ErrorOption {
	return SetData(logOnCreation, true)
}

func CaptureStack() ErrorOption {
	return SetData(captureStack, true)
}

func NoLogOnCreation() ErrorOption {
	return SetData(logOnCreation, false)
}

func NoCaptureStack() ErrorOption {
	return SetData(captureStack, false)
}

func DisableInheritance() ErrorOption {
	return SetData(disableInheritance, true)
}

func boolWrapper(val interface{}, default_value bool) bool {
	rv, ok := val.(bool)
	if ok {
		return rv
	}
	return default_value
}

// NewClass creates an error class with the provided name and options.
func NewClass(name string, options ...ErrorOption) *ErrorClass {
	return HierarchicalError.NewClass(name, options...)
}

// New is for compatibility with the default Go errors package.
func New(text string) error {
	// NewWith doesn't take a format string, even though we have no options.
	return HierarchicalError.NewWith(text)
}

func (parent *ErrorClass) NewClass(name string,
	options ...ErrorOption) *ErrorClass {

	ec := &ErrorClass{
		parent: parent,
		name:   name,
		data:   make(map[DataKey]interface{})}

	for _, option := range options {
		option(ec.data)
	}

	if !boolWrapper(ec.data[disableInheritance], false) {
		// hoist options for speed
		for key, val := range parent.data {
			_, exists := ec.data[key]
			if !exists {
				ec.data[key] = val
			}
		}
		return ec
	} else {
		delete(ec.data, disableInheritance)
	}

	return ec
}

func (e *ErrorClass) Parent() *ErrorClass {
	return e.parent
}

func (e *ErrorClass) String() string {
	return e.name
}

func (e *ErrorClass) Is(parent *ErrorClass) bool {
	for check := e; check != nil; check = check.parent {
		if check == parent {
			return true
		}
	}
	return false
}

// frame logs the pc at some point during execution.
type frame struct {
	pc uintptr
}

// String returns a human readable form of the frame.
func (e frame) String() string {
	if e.pc == 0 {
		return "unknown.unknown:0"
	}
	f := runtime.FuncForPC(e.pc)
	if f == nil {
		return "unknown.unknown:0"
	}
	file, line := f.FileLine(e.pc)
	return fmt.Sprintf("%s:%s:%d", f.Name(), filepath.Base(file), line)
}

// callerState records the pc into an frame for two callers up.
func callerState(depth int) frame {
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		return frame{pc: 0}
	}
	return frame{pc: pc}
}

// record will record the pc at the given depth into the error if it is
// capable of recording it.
func record(err error, depth int) error {
	if err == nil {
		return nil
	}
	cast, ok := err.(*Error)
	if !ok {
		return err
	}
	cast.exits = append(cast.exits, callerState(depth))
	return cast
}

// Record will record the pc of where it is called on to the error.
func Record(err error) error {
	return record(err, 3)
}

// RecordBefore will record the pc depth frames above of where it is called on
// to the error. Record(err) is equivalent to RecordBefore(err, 0)
func RecordBefore(err error, depth int) error {
	return record(err, 3+depth)
}

type Error struct {
	err   error
	class *ErrorClass
	stack []frame
	exits []frame
	data  map[DataKey]interface{}
}

func (e *Error) GetData(key DataKey) interface{} {
	if e.data != nil {
		val, ok := e.data[key]
		if ok {
			return val
		}
		if boolWrapper(e.data[disableInheritance], false) {
			return nil
		}
	}
	return e.class.data[key]
}

func GetData(err error, key DataKey) interface{} {
	cast, ok := err.(*Error)
	if ok {
		return cast.GetData(key)
	}
	return nil
}

func (e *ErrorClass) wrap(err error, classes []*ErrorClass,
	options []ErrorOption) error {
	if err == nil {
		return nil
	}
	if ec, ok := err.(*Error); ok {
		if ec.Is(e) {
			if len(options) == 0 {
				return ec
			}
			// if we have options, we have to wrap it cause we don't want to
			// mutate the existing error.
		} else {
			for _, class := range classes {
				if ec.Is(class) {
					return err
				}
			}
		}
	}

	rv := &Error{err: err, class: e}
	if len(options) > 0 {
		rv.data = make(map[DataKey]interface{})
		for _, option := range options {
			option(rv.data)
		}
	}

	if boolWrapper(rv.GetData(captureStack), false) {
		var pcs [256]uintptr
		amount := runtime.Callers(3, pcs[:])
		rv.stack = make([]frame, amount)
		for i := 0; i < amount; i++ {
			rv.stack[i] = frame{pcs[i]}
		}
	}
	if boolWrapper(rv.GetData(logOnCreation), false) {
		LogWithStack(rv.Error())
	}
	return rv
}

func (e *ErrorClass) WrapUnless(err error, classes ...*ErrorClass) error {
	return e.wrap(err, classes, nil)
}

func (e *ErrorClass) Wrap(err error, options ...ErrorOption) error {
	return e.wrap(err, nil, options)
}

func (e *ErrorClass) New(format string, args ...interface{}) error {
	return e.wrap(fmt.Errorf(format, args...), nil, nil)
}

func (e *ErrorClass) NewWith(message string, options ...ErrorOption) error {
	return e.wrap(errors.New(message), nil, options)
}

func (e *Error) Error() string {
	message := strings.TrimRight(e.err.Error(), "\n ")
	if strings.Contains(message, "\n") {
		message = fmt.Sprintf("%s:\n  %s", e.class.String(),
			strings.Replace(message, "\n", "\n  ", -1))
	} else {
		message = fmt.Sprintf("%s: %s", e.class.String(), message)
	}
	if stack := e.Stack(); stack != "" {
		message = fmt.Sprintf(
			"%s\n\"%s\" backtrace:\n%s", message, e.class, stack)
	}
	if exits := e.Exits(); exits != "" {
		message = fmt.Sprintf(
			"%s\n\"%s\" exits:\n%s", message, e.class, exits)
	}
	return message
}

func (e *Error) Message() string {
	message := strings.TrimRight(GetMessage(e.err), "\n ")
	if strings.Contains(message, "\n") {
		return fmt.Sprintf("%s:\n  %s", e.class.String(),
			strings.Replace(message, "\n", "\n  ", -1))
	}
	return fmt.Sprintf("%s: %s", e.class.String(), message)
}

func (e *Error) WrappedErr() error {
	return e.err
}

func (e *Error) Class() *ErrorClass {
	return e.class
}

func (e *Error) Stack() string {
	if len(e.stack) > 0 {
		frames := make([]string, len(e.stack))
		for i, f := range e.stack {
			frames[i] = f.String()
		}
		return strings.Join(frames, "\n")
	}
	return ""
}

func (e *Error) Exits() string {
	if len(e.exits) > 0 {
		exits := make([]string, len(e.exits))
		for i, ex := range e.exits {
			exits[i] = ex.String()
		}
		return strings.Join(exits, "\n")
	}
	return ""
}

func WrappedErr(err error) error {
	cast, ok := err.(*Error)
	if !ok {
		return err
	}
	return cast.WrappedErr()
}

func GetClass(err error) *ErrorClass {
	if err == nil {
		return nil
	}
	cast, ok := err.(*Error)
	if !ok {
		return findSystemErrorClass(err)
	}
	return cast.class
}

func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	cast, ok := err.(*Error)
	if !ok {
		return err.Error()
	}
	return cast.Message()
}

type EquivalenceOption int

const (
	IncludeWrapped EquivalenceOption = 1
)

func combineEquivOpts(opts []EquivalenceOption) (rv EquivalenceOption) {
	for _, opt := range opts {
		rv |= opt
	}
	return rv
}

func (e *Error) Is(ec *ErrorClass, opts ...EquivalenceOption) bool {
	return ec.Contains(e, opts...)
}

func (e *ErrorClass) Contains(err error, opts ...EquivalenceOption) bool {
	if err == nil {
		return false
	}
	cast, ok := err.(*Error)
	if !ok {
		return findSystemErrorClass(err).Is(e)
	}
	if cast.class.Is(e) {
		return true
	}
	if combineEquivOpts(opts)&IncludeWrapped == 0 {
		return false
	}
	return e.Contains(cast.err, opts...)
}

var (
	// useful error classes
	NotImplementedError = NewClass("Not Implemented Error", LogOnCreation())
	ProgrammerError     = NewClass("Programmer Error", LogOnCreation())
	PanicError          = NewClass("Panic Error", LogOnCreation())
	ErrorGroupError     = NewClass("Error Group Error")

	// classes we fake

	// from os
	SyscallError = SystemError.NewClass("Syscall Error")

	// from syscall
	ErrnoError = SystemError.NewClass("Errno Error")

	// from net
	NetworkError        = SystemError.NewClass("Network Error")
	UnknownNetworkError = NetworkError.NewClass("Unknown Network Error")
	AddrError           = NetworkError.NewClass("Addr Error")
	InvalidAddrError    = AddrError.NewClass("Invalid Addr Error")
	NetOpError          = NetworkError.NewClass("Network Op Error")
	NetParseError       = NetworkError.NewClass("Network Parse Error")
	DNSError            = NetworkError.NewClass("DNS Error")
	DNSConfigError      = DNSError.NewClass("DNS Config Error")

	// from io
	IOError            = SystemError.NewClass("IO Error")
	EOF                = IOError.NewClass("EOF")
	ClosedPipeError    = IOError.NewClass("Closed Pipe Error")
	NoProgressError    = IOError.NewClass("No Progress Error")
	ShortBufferError   = IOError.NewClass("Short Buffer Error")
	ShortWriteError    = IOError.NewClass("Short Write Error")
	UnexpectedEOFError = IOError.NewClass("Unexpected EOF Error")
)

func findSystemErrorClass(err error) *ErrorClass {
	switch err {
	case io.EOF:
		return EOF
	case io.ErrUnexpectedEOF:
		return UnexpectedEOFError
	case io.ErrClosedPipe:
		return ClosedPipeError
	case io.ErrNoProgress:
		return NoProgressError
	case io.ErrShortBuffer:
		return ShortBufferError
	case io.ErrShortWrite:
		return ShortWriteError
	default:
		break
	}
	switch err.(type) {
	case *os.SyscallError:
		return SyscallError
	case syscall.Errno:
		return ErrnoError
	case net.UnknownNetworkError:
		return UnknownNetworkError
	case *net.AddrError:
		return AddrError
	case net.InvalidAddrError:
		return InvalidAddrError
	case *net.OpError:
		return NetOpError
	case *net.ParseError:
		return NetParseError
	case *net.DNSError:
		return DNSError
	case *net.DNSConfigError:
		return DNSConfigError
	case net.Error:
		return NetworkError
	default:
		return SystemError
	}
}
