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
		t.Fatal("expecting nothing to log, got %s", logmsg)
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
	expected = "- foo: BAD\n"
	if !strings.HasSuffix(actual, expected) {
		t.Fatalf("expected suffix %q, got %q", expected, actual)
	}
}
