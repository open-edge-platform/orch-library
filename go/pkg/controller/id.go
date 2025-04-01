// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"fmt"
)

// NewID creates a new object identifier
func NewID(value interface{}) ID {
	return ID{value}
}

// ID is an object identifier
type ID struct {
	// Value is the raw identifier
	Value interface{}
}

// Int returns the identifier as an integer
func (i ID) Int() int {
	return i.Value.(int)
}

// Int32 returns the identifier as an integer
func (i ID) Int32() int32 {
	return i.Value.(int32)
}

// Int64 returns the identifier as an integer
func (i ID) Int64() int64 {
	return i.Value.(int64)
}

// Uint returns the identifier as an integer
func (i ID) Uint() uint {
	return i.Value.(uint)
}

// Uint32 returns the identifier as an integer
func (i ID) Uint32() uint32 {
	return i.Value.(uint32)
}

// Uint64 returns the identifier as an integer
func (i ID) Uint64() uint64 {
	return i.Value.(uint64)
}

// String returns the identifier as a string
func (i ID) String() string {
	return fmt.Sprint(i.Value)
}
