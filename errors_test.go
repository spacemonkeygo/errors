package errors

import (
    "testing"
)

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
