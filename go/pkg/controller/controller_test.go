// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"github.com/open-edge-platform/orch-library/go/pkg/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestController(t *testing.T) {
	ctrl := gomock.NewController(t)

	activatorValue := &atomic.Value{}
	activator := NewMockActivator(ctrl)
	activator.EXPECT().
		Start(gomock.Any()).
		DoAndReturn(func(ch chan<- bool) error {
			activatorValue.Store(ch)
			return nil
		})
	activator.EXPECT().Stop()

	filter := NewMockFilter(ctrl)
	filter.EXPECT().
		Accept(gomock.Any()).
		DoAndReturn(func(id ID) bool {
			i := id.Int()
			return i%2 == 0
		}).
		AnyTimes()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	watcherValue := &atomic.Value{}
	watcher := NewMockWatcher(ctrl)
	watcher.EXPECT().
		Start(gomock.Any()).
		DoAndReturn(func(ch chan<- ID) error {
			watcherValue.Store(ch)
			wg.Done()
			return nil
		})
	watcher.EXPECT().Stop()

	partitions := 3
	partitioner := NewMockWorkPartitioner(ctrl)
	partitioner.EXPECT().
		Partition(gomock.Any()).
		DoAndReturn(func(id ID) (PartitionKey, error) {
			i := id.Int()
			partition := i % partitions
			return PartitionKey(strconv.Itoa(partition)), nil
		}).
		AnyTimes()

	reconciler := NewMockReconciler(ctrl)

	controller := NewController("Test").
		Activate(activator).
		Filter(filter).
		Watch(watcher).
		Partition(partitioner).
		Reconcile(reconciler)
	controller.maxRetryDelay = 100 * time.Millisecond
	defer controller.Stop()

	err := controller.Start()
	assert.NoError(t, err)

	activatorCh := activatorValue.Load().(chan<- bool)
	activatorCh <- true
	done := make(chan struct{})
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(2))).
		Return(Result{}, nil).AnyTimes()
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(4))).
		Return(Result{}, errors.NewInvalid("some error")).Times(4)
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(4))).
		DoAndReturn(func(id ID) (Result, error) {
			return Result{Requeue: id}, nil
		}).Times(3)
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(4))).
		Return(Result{}, errors.NewInvalid("some other error")).Times(4)
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(4))).
		DoAndReturn(func(_ ID) (Result, error) {
			close(done)
			return Result{}, nil
		})
	reconciler.EXPECT().
		Reconcile(gomock.Eq(NewID(4))).
		Return(Result{}, errors.NewInvalid("some error")).AnyTimes()

	wg.Wait()
	watcherCh := watcherValue.Load().(chan<- ID)
	watcherCh <- NewID(1)
	watcherCh <- NewID(2)
	watcherCh <- NewID(3)
	watcherCh <- NewID(4)
	<-done
	watcher.Stop()

}
