// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zerolog

import (
	"bytes"
	"encoding/json"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONEncoder(t *testing.T) {
	framework := &Framework{}
	encoder := framework.JSONEncoder()
	buf := &bytes.Buffer{}
	writer, err := encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Debug("Hello world!")
	assert.Equal(t, "{\"level\":\"debug\",\"message\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.MessageKeyEncoder).WithMessageKey("msg")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"level\":\"info\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameEncoder).WithNameEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Warn("Hello world!")
	assert.Equal(t, "{\"level\":\"warn\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	writer = writer.WithName("test")
	writer.Error("Hello world!")
	assert.Equal(t, "{\"level\":\"error\",\"logger\":\"test\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameKeyEncoder).WithNameKey("name")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer = writer.WithName("test")
	writer.Info("Hello world!")
	assert.Equal(t, "{\"level\":\"info\",\"name\":\"test\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelEncoder).WithLevelEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"level\":\"info\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelKeyEncoder).WithLevelKey("lvl")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"info\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelFormattingEncoder).WithLevelFormat(dazl.UpperCaseLevelFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerEncoder).WithCallerEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Warn("Hello world!")
	assert.Equal(t, "{\"lvl\":\"WARN\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerKeyEncoder).WithCallerKey("call")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"call\":\"encoder_test.go:91\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerKeyEncoder).WithCallerKey("call")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"call\":\"encoder_test.go:99\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerFormattingEncoder).WithCallerFormat(dazl.ShortCallerFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"call\":\"encoder_test.go:107\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerFormattingEncoder).WithCallerFormat(dazl.FullCallerFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assertHasJSONKey(t, "call", buf.Bytes())
	buf.Reset()

	encoder, err = encoder.(dazl.TimestampEncoder).WithTimestampEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assertHasJSONKey(t, "time", buf.Bytes())
	buf.Reset()

	encoder, err = encoder.(dazl.TimestampKeyEncoder).WithTimestampKey("ts")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assertHasJSONKey(t, "ts", buf.Bytes())
	buf.Reset()

	encoder, err = encoder.(dazl.TimestampFormattingEncoder).WithTimestampFormat(dazl.UnixTimestampFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assertHasJSONKey(t, "ts", buf.Bytes())
	buf.Reset()

	encoder, err = encoder.(dazl.TimestampFormattingEncoder).WithTimestampFormat(dazl.ISO8601TimestampFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assertHasJSONKey(t, "ts", buf.Bytes())
	buf.Reset()
}

func TestConsoleEncoder(t *testing.T) {
	framework := &Framework{}
	encoder := framework.ConsoleEncoder()
	buf := &bytes.Buffer{}
	writer, err := encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Debug("Hello world!")
	assert.Equal(t, "\x1b[33mDBG\x1b[0m Hello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameEncoder).WithNameEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "\x1b[32mINF\x1b[0m Hello world!\n", buf.String())
	buf.Reset()

	writer = writer.WithName("test")
	writer.Info("Hello world!")
	assert.Equal(t, "\x1b[32mINF\x1b[0m Hello world! \x1b[36mlogger=\x1b[0mtest\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelEncoder).WithLevelEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Warn("Hello world!")
	assert.Equal(t, "warn Hello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelFormattingEncoder).WithLevelFormat(dazl.UpperCaseLevelFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Error("Hello world!")
	assert.Equal(t, "ERROR Hello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerEncoder).WithCallerEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "INFO \x1b[1mencoder_test.go:195\x1b[0m\x1b[36m >\x1b[0m Hello world!\n", buf.String())
	buf.Reset()
}

func assertHasJSONKey(t *testing.T, key string, data []byte) bool {
	t.Helper()
	object := make(map[string]any)
	assert.NoError(t, json.Unmarshal(data, &object))
	_, ok := object[key]
	return assert.True(t, ok)
}
