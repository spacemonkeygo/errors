// Copyright (C) 2013 Space Monkey, Inc.

package errors

import (
	"sync/atomic"
)

var (
	lastId int32 = 0
)

type DataKey struct{ id int32 }

func GenSym() DataKey { return DataKey{id: atomic.AddInt32(&lastId, 1)} }
