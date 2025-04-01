// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"io"
	"os"
)

func Register(framework Framework) {
	var config loggingConfig
	if err := load(&config); err != nil {
		panic(err)
	} else if err := configure(framework, config, open); err != nil {
		panic(err)
	}
}

func open(path string) (io.Writer, error) {
	switch path {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	}
}

type Framework interface {
	Name() string
}

type ConsoleEncodingFramework interface {
	ConsoleEncoder() Encoder
}

type JSONEncodingFramework interface {
	JSONEncoder() Encoder
}

type defaultFramework struct{}

func (f *defaultFramework) Name() string {
	return "default"
}
