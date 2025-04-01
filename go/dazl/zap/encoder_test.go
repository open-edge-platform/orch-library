// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zap

import (
	"bytes"
	"encoding/json"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestJSONEncoder(t *testing.T) {
	encoder := newJSONEncoder(zapcore.EncoderConfig{})
	buf := &bytes.Buffer{}
	writer, err := encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Debug("Hello world!")
	assert.Equal(t, "{\"message\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.MessageKeyEncoder).WithMessageKey("msg")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameEncoder).WithNameEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Warn("Hello world!")
	assert.Equal(t, "{\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	writer = writer.WithName("test")
	writer.Error("Hello world!")
	assert.Equal(t, "{\"logger\":\"test\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameKeyEncoder).WithNameKey("name")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer = writer.WithName("test")
	writer.Info("Hello world!")
	assert.Equal(t, "{\"name\":\"test\",\"msg\":\"Hello world!\"}\n", buf.String())
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
	writer.Warn("Hello world!")
	assert.Equal(t, "{\"lvl\":\"warn\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelFormattingEncoder).WithLevelFormat(dazl.UpperCaseLevelFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Error("Hello world!")
	assert.Equal(t, "{\"lvl\":\"ERROR\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerEncoder).WithCallerEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"caller\":\"zap/encoder_test.go:83\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerKeyEncoder).WithCallerKey("call")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"call\":\"zap/encoder_test.go:91\",\"msg\":\"Hello world!\"}\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerFormattingEncoder).WithCallerFormat(dazl.ShortCallerFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "{\"lvl\":\"INFO\",\"call\":\"zap/encoder_test.go:99\",\"msg\":\"Hello world!\"}\n", buf.String())
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

	encoder, err = encoder.(dazl.StacktraceEncoder).WithStacktraceEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Error("Hello world!")
	assertHasJSONKey(t, "trace", buf.Bytes())
	buf.Reset()

	encoder, err = encoder.(dazl.StacktraceKeyEncoder).WithStacktraceKey("stack")
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Error("Hello world!")
	assertHasJSONKey(t, "stack", buf.Bytes())
	buf.Reset()
}

func TestConsoleEncoder(t *testing.T) {
	encoder := newConsoleEncoder(zapcore.EncoderConfig{})
	buf := &bytes.Buffer{}
	writer, err := encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Debug("Hello world!")
	assert.Equal(t, "Hello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.NameEncoder).WithNameEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "Hello world!\n", buf.String())
	buf.Reset()

	writer = writer.WithName("test")
	writer.Warn("Hello world!")
	assert.Equal(t, "test\tHello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelEncoder).WithLevelEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Error("Hello world!")
	assert.Equal(t, "error\tHello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.LevelFormattingEncoder).WithLevelFormat(dazl.UpperCaseLevelFormat)
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Info("Hello world!")
	assert.Equal(t, "INFO\tHello world!\n", buf.String())
	buf.Reset()

	encoder, err = encoder.(dazl.CallerEncoder).WithCallerEnabled()
	assert.NoError(t, err)
	writer, err = encoder.NewWriter(buf)
	assert.NoError(t, err)
	writer.Warn("Hello world!")
	assert.Equal(t, "WARN\tzap/encoder_test.go:202\tHello world!\n", buf.String())
	buf.Reset()
}

func assertHasJSONKey(t *testing.T, key string, data []byte) bool {
	t.Helper()
	object := make(map[string]any)
	assert.NoError(t, json.Unmarshal(data, &object))
	_, ok := object[key]
	return assert.True(t, ok)
}
