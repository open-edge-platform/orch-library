// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUnmarshalSampler(t *testing.T) {
	text := `
basic:
  interval: 10
  maxLevel: info
`
	config := &samplingConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), config))
	assert.NotNil(t, config.Basic)
	assert.Nil(t, config.Random)
	assert.Equal(t, 10, config.Basic.Interval)
	assert.Equal(t, InfoLevel, config.Basic.MaxLevel.Level())

	text = "random"
	config = &samplingConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), config))
	assert.Nil(t, config.Basic)
	assert.NotNil(t, config.Random)
	assert.Equal(t, EmptyLevel, config.Random.MaxLevel.Level())

	text = `
random:
  maxLevel: info`
	config = &samplingConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(text), config))
	assert.Nil(t, config.Basic)
	assert.NotNil(t, config.Random)
	assert.Equal(t, InfoLevel, config.Random.MaxLevel.Level())
}
