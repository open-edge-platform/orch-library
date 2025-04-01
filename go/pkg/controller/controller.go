// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
	"math"
	"sync"
	"time"
)

var log = dazl.GetLogger()

const (
	maxRetryDelayDefault = 5 * time.Second
	delayStep            = 10 * time.Millisecond
)

// Watcher is implemented by controllers to implement watching for specific events
// Type identifiers that are written to the watcher channel will eventually be processed by
// the controller.
type Watcher interface {
	// Start starts watching for events
	Start(ch chan<- ID) error

	// Stop stops watching for events
	Stop()
}

// Reconciler reconciles objects
// The reconciler will be called for each type ID received from the Watcher. The reconciler may indicate
// whether to retry requests by returning either false or a non-nil error. Reconcilers should make no
// assumptions regarding the ordering of requests and should use the provided type IDs to resolve types
// against the current state of the cluster.
type Reconciler interface {
	// Reconcile is called to reconcile the state of an object
	Reconcile(ID) (Result, error)
}

// Request is a reconciler request
type Request struct {
	ID      ID
	attempt int
}

// Result is a reconciler result
type Result struct {
	// Requeue is the identifier of an event to requeue
	Requeue      ID
	RequeueAt    time.Time
	RequeueAfter time.Duration
}

// NewController creates a new controller
func NewController(name string) *Controller {
	return &Controller{
		name:          name,
		activator:     &UnconditionalActivator{},
		partitioner:   &UnaryPartitioner{},
		watchers:      make([]Watcher, 0),
		partitions:    make(map[PartitionKey]chan Request),
		maxRetryDelay: maxRetryDelayDefault,
	}
}

// Controller is a control loop
// The Controller is responsible for processing events provided by a Watcher. Events are processed by
// a configurable Reconciler. The controller processes events in a loop, retrying requests until the
// Reconciler can successfully process them.
// The Controller can be activated or deactivated by a configurable Activator. When inactive, the controller
// will ignore requests, and when active it processes all requests.
// For per-request filtering, a Filter can be provided which provides a simple bool to indicate whether a
// request should be passed to the Reconciler.
// Once the Reconciler receives a request, it should process the request using the current state of the cluster
// Reconcilers should not cache state themselves and should instead rely on stores for consistency.
// If a Reconciler returns false, the request will be requeued to be retried after all pending requests.
// If a Reconciler returns an error, the request will be retried after a backoff period.
// Once a Reconciler successfully processes a request by returning true, the request will be discarded.
// Requests can be partitioned among concurrent goroutines by configuring a WorkPartitioner. The controller
// will create a goroutine per PartitionKey provided by the WorkPartitioner, and requests to different
// partitions may be handled concurrently.
type Controller struct {
	name          string
	mu            sync.RWMutex
	activator     Activator
	partitioner   WorkPartitioner
	filter        Filter
	watchers      []Watcher
	reconciler    Reconciler
	partitions    map[PartitionKey]chan Request
	maxRetryDelay time.Duration
}

// Activate sets an activator for the controller
func (c *Controller) Activate(activator Activator) *Controller {
	c.mu.Lock()
	c.activator = activator
	c.mu.Unlock()
	return c
}

// Partition partitions work among multiple goroutines for the controller
func (c *Controller) Partition(partitioner WorkPartitioner) *Controller {
	c.mu.Lock()
	c.partitioner = partitioner
	c.mu.Unlock()
	return c
}

// Filter sets a filter for the controller
func (c *Controller) Filter(filter Filter) *Controller {
	c.mu.Lock()
	c.filter = filter
	c.mu.Unlock()
	return c
}

// Watch adds a watcher to the controller
func (c *Controller) Watch(watcher Watcher) *Controller {
	c.mu.Lock()
	c.watchers = append(c.watchers, watcher)
	c.mu.Unlock()
	return c
}

// Reconcile sets the reconciler for the controller
func (c *Controller) Reconcile(reconciler Reconciler) *Controller {
	c.mu.Lock()
	c.reconciler = reconciler
	c.mu.Unlock()
	return c
}

// Start starts the request controller
func (c *Controller) Start() error {
	ch := make(chan bool)
	if err := c.activator.Start(ch); err != nil {
		return err
	}
	go func() {
		active := false
		for activate := range ch {
			if activate {
				if !active {
					log.Infof("Activating controller %s", c.name)
					c.activate()
					active = true
				}
			} else {
				if active {
					log.Infof("Deactivating controller %s", c.name)
					c.deactivate()
					active = false
				}
			}
		}
		if active {
			log.Infof("Deactivating controller %s", c.name)
			c.deactivate()
		}
	}()
	return nil
}

// Stop stops the controller
func (c *Controller) Stop() {
	c.activator.Stop()
}

// activate activates the controller
func (c *Controller) activate() {
	ch := make(chan ID)
	wg := &sync.WaitGroup{}
	for _, watcher := range c.watchers {
		if err := c.startWatcher(ch, wg, watcher); err == nil {
			wg.Add(1)
		}
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	go c.processEvents(ch)
}

// startWatcher starts a single watcher
func (c *Controller) startWatcher(ch chan ID, wg *sync.WaitGroup, watcher Watcher) error {
	watcherCh := make(chan ID)
	if err := watcher.Start(watcherCh); err != nil {
		return err
	}

	go func() {
		for id := range watcherCh {
			ch <- id
		}
		wg.Done()
	}()
	return nil
}

// deactivate deactivates the controller
func (c *Controller) deactivate() {
	for _, watcher := range c.watchers {
		watcher.Stop()
	}
}

// processEvents processes the events from the given channel
func (c *Controller) processEvents(ch chan ID) {
	c.mu.RLock()
	filter := c.filter
	partitioner := c.partitioner
	c.mu.RUnlock()
	for id := range ch {
		// Ensure the event is applicable to this controller
		if filter == nil || filter.Accept(id) {
			c.partition(id, partitioner)
		}
	}
}

// partition writes the given request to a partition
func (c *Controller) partition(id ID, partitioner WorkPartitioner) {
	c.mu.RLock()
	reconciler := c.reconciler
	c.mu.RUnlock()

	iteration := 1
	for {
		// Get the partition key for the object ID
		key, err := partitioner.Partition(id)
		if err != nil {
			time.Sleep(delayStep * time.Duration(math.Pow(2, float64(iteration))))
		} else {
			// Get or create a partition channel for the partition key
			c.mu.RLock()
			partition, ok := c.partitions[key]
			c.mu.RUnlock()
			if !ok {
				c.mu.Lock()
				partition, ok = c.partitions[key]
				if !ok {
					partition = make(chan Request)
					c.partitions[key] = partition
					go c.reconcileRequests(partition, reconciler)
				}
				c.mu.Unlock()
			}
			partition <- Request{
				ID: id,
			}
			return
		}
		iteration++
	}
}

// reconcileRequests reconciles requests from the given channel
func (c *Controller) reconcileRequests(ch chan Request, reconciler Reconciler) {
	for request := range ch {
		c.reconcileRequest(request, ch, reconciler)
	}
}

// reconcileRequest reconciles the given request
func (c *Controller) reconcileRequest(request Request, ch chan Request, reconciler Reconciler) {
	request.attempt++
	result, err := reconciler.Reconcile(request.ID)
	if err != nil {
		maxExponent := math.Log2(float64(c.maxRetryDelay) / float64(delayStep))
		retryDelay := delayStep * time.Duration(math.Pow(2, math.Min(float64(request.attempt), maxExponent)))
		log.Infof("error during reconciliation of %v. Attempt %d. Retrying after %s: %s",
			request.ID.Value, request.attempt, retryDelay, err)
		time.AfterFunc(retryDelay, func() {
			ch <- request
		})
	} else if !result.RequeueAt.IsZero() {
		time.AfterFunc(time.Until(result.RequeueAt), func() {
			if result.Requeue.Value != nil {
				ch <- Request{
					ID: result.Requeue,
				}
			} else {
				ch <- Request{
					ID: request.ID,
				}
			}
		})
	} else if result.RequeueAfter > 0 {
		time.AfterFunc(result.RequeueAfter, func() {
			if result.Requeue.Value != nil {
				ch <- Request{
					ID: result.Requeue,
				}
			} else {
				ch <- Request{
					ID: request.ID,
				}
			}
		})
	} else if result.Requeue.Value != nil {
		go func() {
			ch <- Request{
				ID: result.Requeue,
			}
		}()
	}
}
