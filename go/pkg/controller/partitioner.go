// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"regexp"
)

// PartitionKey is a key by which to partition requests
type PartitionKey string

const staticKey PartitionKey = ""

// WorkPartitioner is an interface for partitioning requests among a set of goroutines
// The WorkPartitioner can enable safe concurrency inside controllers. For each request, the
// partitioner will be called to provide a PartitionKey for the request. For each unique
// PartitionKey, a separate channel and goroutine will be created to process requests for
// the partition.
type WorkPartitioner interface {
	// Partition gets a partition key for the given request
	Partition(id ID) (PartitionKey, error)
}

// UnaryPartitioner is a WorkPartitioner that assigns all work to a single goroutine
type UnaryPartitioner struct {
}

// Partition returns a static partition key
func (p *UnaryPartitioner) Partition(_ ID) (PartitionKey, error) {
	return staticKey, nil
}

var _ WorkPartitioner = &UnaryPartitioner{}

// RegexpPartitioner is a WorkPartitioner that assigns work to a gouroutine per regex output
type RegexpPartitioner struct {
	Regexp regexp.Regexp
}

// Partition returns a PartitionKey from the configured regex
func (p *RegexpPartitioner) Partition(id ID) (PartitionKey, error) {
	return PartitionKey(p.Regexp.FindString(id.String())), nil
}

var _ WorkPartitioner = &RegexpPartitioner{}
