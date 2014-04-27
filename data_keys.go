// Copyright (C) 2013 Space Monkey, Inc.

package errors

import (
	"sync/atomic"
)

var (
	lastId int32 = 0
)

// DataKey's job is to make sure that keys in each error instances namespace
// are lexically scoped, thus helping developers not step on each others' toes
// between large packages. You can only store data on an error using a DataKey,
// and you can only make DataKeys with GenSym().
type DataKey struct{ id int32 }

// GenSym generates a brand new, never-before-seen DataKey
func GenSym() DataKey { return DataKey{id: atomic.AddInt32(&lastId, 1)} }
