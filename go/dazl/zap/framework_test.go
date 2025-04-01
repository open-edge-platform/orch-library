// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFramework(t *testing.T) {
	framework := &Framework{}
	assert.Equal(t, "zap", framework.Name())
	assert.NotNil(t, framework.JSONEncoder())
	assert.NotNil(t, framework.ConsoleEncoder())
}
