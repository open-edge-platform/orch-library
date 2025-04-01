// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zap

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
	"go.uber.org/zap/zapcore"
)

func init() {
	dazl.Register(&Framework{})
}

type Framework struct{}

func (f *Framework) Name() string {
	return "zap"
}

func (f *Framework) ConsoleEncoder() dazl.Encoder {
	return newConsoleEncoder(zapcore.EncoderConfig{})
}

func (f *Framework) JSONEncoder() dazl.Encoder {
	return newJSONEncoder(zapcore.EncoderConfig{})
}
