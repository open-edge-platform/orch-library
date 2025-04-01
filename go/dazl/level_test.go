// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestLevel(t *testing.T) {
	assert.False(t, DebugLevel.Enabled(EmptyLevel))
	assert.True(t, DebugLevel.Enabled(DebugLevel))
	assert.False(t, InfoLevel.Enabled(DebugLevel))
	assert.True(t, InfoLevel.Enabled(InfoLevel))
	assert.False(t, WarnLevel.Enabled(InfoLevel))
	assert.True(t, WarnLevel.Enabled(WarnLevel))
	assert.False(t, ErrorLevel.Enabled(WarnLevel))
	assert.True(t, ErrorLevel.Enabled(ErrorLevel))
	assert.False(t, PanicLevel.Enabled(ErrorLevel))
	assert.True(t, PanicLevel.Enabled(PanicLevel))
	assert.False(t, FatalLevel.Enabled(PanicLevel))
	assert.True(t, FatalLevel.Enabled(FatalLevel))
}

const testLevel = "info"

func TestUnmarshalLevel(t *testing.T) {
	var level levelConfig
	assert.NotEqual(t, InfoLevel, level.Level())
	assert.NoError(t, yaml.Unmarshal([]byte(testLevel), &level))
	assert.Equal(t, InfoLevel, level.Level())
}
