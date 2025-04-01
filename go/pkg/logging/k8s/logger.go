// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/open-edge-platform/orch-library/go/dazl"
)

// NewControllerLogger returns a new implementation of the Kubernetes Logger interface
func NewControllerLogger(names ...string) logr.Logger {
	controllerLogger := &ControllerLogger{
		log: dazl.GetLogger(names...),
	}

	return logr.New(controllerLogger)
}

// NewControllerPackageLogger returns a new implementation of the Kubernetes Logger interface
func NewControllerPackageLogger() logr.Logger {
	controllerLogger := &ControllerLogger{
		log: dazl.GetPackageLogger(),
	}
	return logr.New(controllerLogger)
}

func getFields(keysAndValues ...interface{}) []dazl.Field {
	fields := make([]dazl.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key := keysAndValues[i]
		value := keysAndValues[i+1]
		field := dazl.String(fmt.Sprint(key), fmt.Sprint(value))
		fields = append(fields, field)
	}
	return fields
}

// ControllerLogger is an implementation of the Kubernetes controller Logger interface
type ControllerLogger struct {
	log dazl.Logger
}

func (l ControllerLogger) WithCallDepth(depth int) logr.LogSink {
	return &ControllerLogger{
		log: l.log.WithSkipCalls(depth),
	}
}

func (l ControllerLogger) Init(_ logr.RuntimeInfo) {}

func (l ControllerLogger) Enabled(_ int) bool {
	return true
}

func (l ControllerLogger) Info(_ int, msg string, keysAndValues ...interface{}) {
	l.log.WithFields(getFields(keysAndValues...)...).Info(msg)
}

func (l ControllerLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.log.WithFields(getFields(keysAndValues...)...).Error(err, msg)
}

func (l ControllerLogger) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return &ControllerLogger{
		log: l.log.WithFields(getFields(keysAndValues...)...),
	}
}

func (l ControllerLogger) WithName(name string) logr.LogSink {
	return &ControllerLogger{
		log: l.log.GetLogger(name),
	}
}
