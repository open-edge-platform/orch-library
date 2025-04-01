// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zap

import (
	"fmt"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

type zapEncoder struct {
	config  zapcore.EncoderConfig
	newFunc func(zapcore.EncoderConfig) dazl.Encoder
}

func (e *zapEncoder) with(f func(*zapcore.EncoderConfig)) dazl.Encoder {
	config := e.config
	f(&config)
	return e.newFunc(config)
}

func (e *zapEncoder) WithNameEnabled() (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.NameKey = "logger"
	}), nil
}

func (e *zapEncoder) WithLevelEnabled() (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.LevelKey = "level"
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}), nil
}

func (e *zapEncoder) WithLevelFormat(format dazl.LevelFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.LowerCaseLevelFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeLevel = zapcore.LowercaseLevelEncoder
		}), nil
	case dazl.UpperCaseLevelFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeLevel = zapcore.CapitalLevelEncoder
		}), nil
	default:
		return nil, fmt.Errorf("unsupported level format '%s'", format)
	}
}

func (e *zapEncoder) WithTimestampEnabled() (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.TimeKey = "time"
		config.EncodeTime = zapcore.EpochTimeEncoder
	}), nil
}

func (e *zapEncoder) WithTimestampFormat(format dazl.TimestampFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.ISO8601TimestampFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeTime = zapcore.ISO8601TimeEncoder
		}), nil
	case dazl.UnixTimestampFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeTime = zapcore.EpochTimeEncoder
		}), nil
	default:
		return nil, fmt.Errorf("unsupported time format '%s'", format)
	}
}

func (e *zapEncoder) WithCallerEnabled() (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.CallerKey = "caller"
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}), nil
}

func (e *zapEncoder) WithCallerFormat(format dazl.CallerFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.ShortCallerFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeCaller = zapcore.ShortCallerEncoder
		}), nil
	case dazl.FullCallerFormat:
		return e.with(func(config *zapcore.EncoderConfig) {
			config.EncodeCaller = zapcore.FullCallerEncoder
		}), nil
	default:
		return nil, fmt.Errorf("unsupported caller format '%s'", format)
	}
}

func (e *zapEncoder) WithStacktraceEnabled() (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.StacktraceKey = "trace"
	}), nil
}

func newConsoleEncoder(config zapcore.EncoderConfig) dazl.Encoder {
	return &consoleEncoder{
		zapEncoder: &zapEncoder{
			config:  config,
			newFunc: newConsoleEncoder,
		},
	}
}

type consoleEncoder struct {
	*zapEncoder
}

func (e *consoleEncoder) NewWriter(writer io.Writer) (dazl.Writer, error) {
	if e.config.MessageKey == "" {
		e.config.MessageKey = "message"
	}
	var config zap.Config
	config.EncoderConfig = e.config
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	return newWriter(writer, zapcore.NewConsoleEncoder(e.config), config)
}

var _ dazl.NameEncoder = (*consoleEncoder)(nil)
var _ dazl.LevelEncoder = (*consoleEncoder)(nil)
var _ dazl.LevelFormattingEncoder = (*consoleEncoder)(nil)
var _ dazl.TimestampEncoder = (*consoleEncoder)(nil)
var _ dazl.TimestampFormattingEncoder = (*consoleEncoder)(nil)
var _ dazl.CallerEncoder = (*consoleEncoder)(nil)
var _ dazl.CallerFormattingEncoder = (*consoleEncoder)(nil)
var _ dazl.StacktraceEncoder = (*consoleEncoder)(nil)

func newJSONEncoder(config zapcore.EncoderConfig) dazl.Encoder {
	return &jsonEncoder{
		zapEncoder: &zapEncoder{
			config:  config,
			newFunc: newJSONEncoder,
		},
	}
}

type jsonEncoder struct {
	*zapEncoder
}

func (e *jsonEncoder) NewWriter(writer io.Writer) (dazl.Writer, error) {
	if e.config.MessageKey == "" {
		e.config.MessageKey = "message"
	}
	var config zap.Config
	config.EncoderConfig = e.config
	config.Encoding = "json"
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	return newWriter(writer, zapcore.NewJSONEncoder(e.config), config)
}

func (e *jsonEncoder) WithMessageKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.MessageKey = key
	}), nil
}

func (e *jsonEncoder) WithNameKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.NameKey = key
	}), nil
}

func (e *jsonEncoder) WithLevelKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.LevelKey = key
	}), nil
}

func (e *jsonEncoder) WithTimestampKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.TimeKey = key
	}), nil
}

func (e *jsonEncoder) WithCallerKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.CallerKey = key
	}), nil
}

func (e *jsonEncoder) WithStacktraceKey(key string) (dazl.Encoder, error) {
	return e.with(func(config *zapcore.EncoderConfig) {
		config.StacktraceKey = key
	}), nil
}

var _ dazl.MessageKeyEncoder = (*jsonEncoder)(nil)
var _ dazl.NameEncoder = (*jsonEncoder)(nil)
var _ dazl.NameKeyEncoder = (*jsonEncoder)(nil)
var _ dazl.LevelEncoder = (*jsonEncoder)(nil)
var _ dazl.LevelKeyEncoder = (*jsonEncoder)(nil)
var _ dazl.LevelFormattingEncoder = (*jsonEncoder)(nil)
var _ dazl.TimestampEncoder = (*jsonEncoder)(nil)
var _ dazl.TimestampKeyEncoder = (*jsonEncoder)(nil)
var _ dazl.TimestampFormattingEncoder = (*jsonEncoder)(nil)
var _ dazl.CallerEncoder = (*jsonEncoder)(nil)
var _ dazl.CallerKeyEncoder = (*jsonEncoder)(nil)
var _ dazl.CallerFormattingEncoder = (*jsonEncoder)(nil)
var _ dazl.StacktraceEncoder = (*jsonEncoder)(nil)
var _ dazl.StacktraceKeyEncoder = (*jsonEncoder)(nil)
