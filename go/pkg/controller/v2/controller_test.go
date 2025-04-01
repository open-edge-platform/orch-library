// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

type testID string

func (id testID) String() string {
	return string(id)
}

func TestAck(t *testing.T) {
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("foo"))
	<-done
}

func TestFail(t *testing.T) {
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		close(done)
		return request.Fail(errors.New("test"))
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("foo"))
	<-done
}

func TestRequeue(t *testing.T) {
	var requeued atomic.Bool
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		if requeued.CompareAndSwap(false, true) {
			return request.Requeue()
		}
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("bar"))
	<-done
}

func TestRetry(t *testing.T) {
	var retried atomic.Bool
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		if retried.CompareAndSwap(false, true) {
			return request.Retry(errors.New("test"))
		}
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("baz"))
	<-done
}

func TestRetryAfter(t *testing.T) {
	var retried atomic.Bool
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		if retried.CompareAndSwap(false, true) {
			return request.Retry(errors.New("test")).After(time.Second)
		}
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("foo"))
	<-done
}

func TestRetryAt(t *testing.T) {
	var retried atomic.Bool
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		if retried.CompareAndSwap(false, true) {
			return request.Retry(errors.New("test")).At(time.Now().Add(time.Second))
		}
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("bar"))
	<-done
}

func TestRetryWith(t *testing.T) {
	var retried atomic.Bool
	done := make(chan struct{})
	controller := NewController[testID](func(_ context.Context, request Request[testID]) Directive[testID] {
		if retried.CompareAndSwap(false, true) {
			return request.Retry(errors.New("test")).With(ExponentialBackoff(time.Second, time.Minute))
		}
		close(done)
		return request.Ack()
	}, WithParallelism(10))
	defer controller.Stop()
	assert.NoError(t, controller.Reconcile("bar"))
	<-done
}
