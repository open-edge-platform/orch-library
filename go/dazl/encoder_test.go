// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

const testEncoders = `
console:
  fields:
    message:
    level:
      key: the-level
    timestamp:
      format: unix
    caller:
json:
  fields:
    - message
    - level:
        format: lowercase
    - timestamp:
        key: timestamp
    - caller:
        format: full
`

func TestUnmarshalEncoders(t *testing.T) {
	var encoders encodersConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testEncoders), &encoders))

	assert.NotNil(t, encoders.Console)
	assert.NotNil(t, encoders.Console.Fields.Message)
	assert.NotNil(t, encoders.Console.Fields.Level)
	assert.NotNil(t, encoders.Console.Fields.Time)
	assert.NotNil(t, encoders.Console.Fields.Caller)
	assert.Nil(t, encoders.Console.Fields.Stacktrace)

	assert.Equal(t, "", encoders.Console.Fields.Message.Key)
	assert.Equal(t, "the-level", encoders.Console.Fields.Level.Key)
	assert.Equal(t, "", encoders.Console.Fields.Time.Key)
	assert.Equal(t, "", encoders.Console.Fields.Caller.Key)

	assert.Nil(t, encoders.Console.Fields.Level.Format)
	assert.NotNil(t, encoders.Console.Fields.Time.Format)
	assert.Nil(t, encoders.Console.Fields.Caller.Format)
	assert.Equal(t, UnixTimestampFormat, *encoders.Console.Fields.Time.Format)

	assert.NotNil(t, encoders.JSON)
	assert.NotNil(t, encoders.JSON.Fields.Message)
	assert.NotNil(t, encoders.JSON.Fields.Level)
	assert.NotNil(t, encoders.JSON.Fields.Time)
	assert.NotNil(t, encoders.JSON.Fields.Caller)
	assert.Nil(t, encoders.JSON.Fields.Stacktrace)

	assert.Equal(t, "", encoders.JSON.Fields.Message.Key)
	assert.Equal(t, "", encoders.JSON.Fields.Level.Key)
	assert.Equal(t, "timestamp", encoders.JSON.Fields.Time.Key)
	assert.Equal(t, "", encoders.JSON.Fields.Caller.Key)

	assert.NotNil(t, encoders.JSON.Fields.Level.Format)
	assert.Nil(t, encoders.JSON.Fields.Time.Format)
	assert.NotNil(t, encoders.JSON.Fields.Caller.Format)
	assert.Equal(t, LowerCaseLevelFormat, *encoders.JSON.Fields.Level.Format)
	assert.Equal(t, FullCallerFormat, *encoders.JSON.Fields.Caller.Format)
}
