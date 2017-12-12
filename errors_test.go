// Copyright (C) 2014 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

var (
	logbuf = new(bytes.Buffer)
)

func init() {
	log.SetFlags(0)
	log.SetOutput(logbuf)
}

func testRecord0() error {
	return Record(testRecord1())
}

func testRecord1() error {
	return Record(testRecord2())
}

func testRecord2() error {
	return HierarchicalError.New("testing")
}

func TestRecord(t *testing.T) {
	t.Log(testRecord0())
}

func TestBacktrace(t *testing.T) {
	t.Log(testRecord0())
	ch := make(chan bool)
	go func() {
		t.Log(testRecord0())
		ch <- true
	}()
	<-ch
}

func TestErrorGroupReturnsNilIfNoneAdded(t *testing.T) {
	errs := NewErrorGroup()
	err := errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}
	errs.Add(nil)
	err = errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}
}

func TestErrorGroupLimitReturnsNilIfNoneAdded(t *testing.T) {
	errs := NewBoundedErrorGroup(1)
	err := errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}
	errs.Add(nil)
	err = errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}
}

func TestErrorGroupLimitNotExceeded(t *testing.T) {
	errs := NewBoundedErrorGroup(1)
	errs.Add(fmt.Errorf("BAD"))
	actual := errs.Finalize().Error()
	if actual != "BAD" {
		t.Fatalf(`expected "BAD", got %s`, actual)
	}
}

func TestErrorGroupLimitExceeded(t *testing.T) {
	errs := NewBoundedErrorGroup(1)
	errs.Add(fmt.Errorf("BAD"))
	errs.Add(fmt.Errorf("MAD"))
	errs.Add(fmt.Errorf("SAD"))
	actual := errs.Finalize().Error()
	expected := `Error Group Error:
  BAD
  ... and 2 more.`
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf(`expected prefix %q, got %q`, expected, actual)
	}
}

func TestLoggingErrorGroupReturnsNilIfNoneAdded(t *testing.T) {
	logbuf.Reset()

	errs := NewLoggingErrorGroup("foo")
	err := errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}
	errs.Add(nil)
	err = errs.Finalize()
	if err != nil {
		t.Fatalf("expected nil, got %s", err)
	}

	logmsg := string(logbuf.Bytes())
	if logmsg != "" {
		t.Fatalf("expecting nothing to log, got %s", logmsg)
	}
}

func TestLoggingErrorGroupLogsAndReturnsErrIfAdded(t *testing.T) {
	logbuf.Reset()

	errs := NewLoggingErrorGroup("foo")
	errs.Add(fmt.Errorf("BAD"))
	errs.Add(nil)
	actual := errs.Finalize().Error()
	expected := "Error Group Error: foo: 1 of 2 failed."
	if !strings.HasPrefix(actual, expected) {
		t.Fatalf("expected prefix %q, got %q", expected, actual)
	}

	actual = string(logbuf.Bytes())
	expected = "foo: BAD\n"
	if !strings.HasSuffix(actual, expected) {
		t.Fatalf("expected suffix %q, got %q", expected, actual)
	}
}

func TestErrorName(t *testing.T) {
	name, ok := HierarchicalError.New("test").(*Error).Name()
	assert(t, ok)
	assert(t, name == "Error")
}

func assert(t *testing.T, val bool) {
	if !val {
		t.Fatal("assertion failed")
	}
}

func ExampleSetData(t *testing.T) {
	// Create our own DataKeys
	UserMessageKey := GenSym()
	ConstraintNameKey := GenSym()
	OtherKey := GenSym()

	// Create some error classes
	ApplicationError := NewClass("Application Error")
	ConstraintError := ApplicationError.NewClass("Constraint Error",
		SetData(UserMessageKey, "A constraint failed on your data"))
	ValueConstraintError := ConstraintError.NewClass("Value Constraint Error")

	// Create an actual error. Something bad happened.
	err := ValueConstraintError.NewWith("value constraint failed!",
		SetData(ConstraintNameKey, "equality_constraint"))

	// Make sure everything is how we expect
	assert(t, ValueConstraintError.Contains(err))
	assert(t, ConstraintError.Contains(err))
	assert(t, ApplicationError.Contains(err))
	assert(t, !SystemError.Contains(err))

	assert(t, GetData(err, UserMessageKey).(string) ==
		"A constraint failed on your data")
	assert(t, GetData(err, ConstraintNameKey).(string) == "equality_constraint")
	assert(t, GetData(err, OtherKey) == nil)
}

func ExampleCatchPanic(t *testing.T) {
	panicfn := func() {
		panic("oh hai")
	}
	nonpanicfn := func() (err error) {
		defer CatchPanic(&err)
		panicfn()
		return nil
	}
	err := nonpanicfn()
	assert(t, PanicError.Contains(err))
}

func ExampleErrorGroup(t *testing.T) {
	// example utils
	work := func(i int) error {
		if i%2 == 0 {
			return nil
		}
		return New("error")
	}
	handle_err := func(error) {}

	// example:
	errs := NewErrorGroup()
	for i := 0; i < 10; i++ {
		errs.Add(work(i))
	}
	err := errs.Finalize()
	if err != nil {
		handle_err(err)
	}
}
