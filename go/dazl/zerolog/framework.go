// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zerolog

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
)

func init() {
	dazl.Register(&Framework{})
}

type Framework struct{}

func (f *Framework) Name() string {
	return "zerolog"
}

func (f *Framework) ConsoleEncoder() dazl.Encoder {
	return &consoleEncoder{}
}

func (f *Framework) JSONEncoder() dazl.Encoder {
	return &jsonEncoder{}
}
