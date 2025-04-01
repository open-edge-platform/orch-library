// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

// Activator is an interface for controlling the activation of a controller
// Once the Activator is Started, it may activate or deactivate processing of Watcher events on the
// node at any time by writing true or false to the activator channel respectively.
type Activator interface {
	// Start starts the activator
	Start(ch chan<- bool) error

	// Stop stops the activator
	Stop()
}

// UnconditionalActivator activates controllers on all nodes at all times
type UnconditionalActivator struct {
	ch chan<- bool
}

// Start starts the activator
func (a *UnconditionalActivator) Start(ch chan<- bool) error {
	a.ch = ch
	go func() {
		ch <- true
	}()
	return nil
}

// Stop stops the activator
func (a *UnconditionalActivator) Stop() {
	if a.ch != nil {
		close(a.ch)
	}
}

var _ Activator = &UnconditionalActivator{}
