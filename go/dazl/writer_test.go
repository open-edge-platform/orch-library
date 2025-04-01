// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

const testWriters = `
stdout:
  encoder: console
stderr:
  encoder: console
file:
  path: ./foo/bar
  encoder: json
`

func TestUnmarshalWriters(t *testing.T) {
	var writers writersConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testWriters), &writers))

	assert.NotNil(t, writers.Stdout)
	assert.NotNil(t, writers.Stderr)
	assert.Len(t, writers.Files, 1)

	assert.Equal(t, ConsoleEncoding, writers.Stdout.Encoder)
	assert.Equal(t, ConsoleEncoding, writers.Stderr.Encoder)
	assert.Equal(t, JSONEncoding, writers.Files["file"].Encoder)
	assert.Equal(t, "./foo/bar", writers.Files["file"].Path)
}
