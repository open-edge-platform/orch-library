// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zerolog

import (
	"fmt"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/rs/zerolog"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type consoleEncoder struct {
	writer     zerolog.ConsoleWriter
	nameKey    string
	timestamp  bool
	caller     bool
	stacktrace bool
}

func (e *consoleEncoder) NewWriter(out io.Writer) (dazl.Writer, error) {
	writer := e.writer
	writer.Out = out
	if !e.timestamp {
		writer.FormatTimestamp = func(_ any) string {
			return ""
		}
	}
	logger := zerolog.New(writer)
	logger = logger.With().Timestamp().Logger()
	if e.caller {
		logger = logger.With().CallerWithSkipFrameCount(3).Logger()
	}
	if e.stacktrace {
		logger = logger.With().Stack().Logger()
	}
	return &Writer{
		logger:     logger,
		nameKey:    e.nameKey,
		skipFrames: 3,
	}, nil
}

func (e *consoleEncoder) WithNameEnabled() (dazl.Encoder, error) {
	return &consoleEncoder{
		nameKey: "logger",
	}, nil
}

func (e *consoleEncoder) WithNameKey(key string) (dazl.Encoder, error) {
	return &consoleEncoder{
		nameKey: key,
	}, nil
}

func (e *consoleEncoder) WithLevelEnabled() (dazl.Encoder, error) {
	return e.WithLevelFormat(dazl.LowerCaseLevelFormat)
}

func (e *consoleEncoder) WithLevelFormat(format dazl.LevelFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.LowerCaseLevelFormat:
		e.writer.FormatLevel = func(level interface{}) string {
			return strings.ToLower(fmt.Sprint(level))
		}
		return e, nil
	case dazl.UpperCaseLevelFormat:
		e.writer.FormatLevel = func(level interface{}) string {
			return strings.ToUpper(fmt.Sprint(level))
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unsupported level format '%s'", format)
	}
}

func (e *consoleEncoder) WithTimestampEnabled() (dazl.Encoder, error) {
	e.timestamp = true
	return e, nil
}

func (e *consoleEncoder) WithTimestampFormat(format dazl.TimestampFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.UnixTimestampFormat:
		e.writer.TimeFormat = zerolog.TimeFormatUnix
	case dazl.ISO8601TimestampFormat:
		e.writer.TimeFormat = time.RFC3339
	default:
		return nil, fmt.Errorf("unsupoorted timestamp format %s", format)
	}
	return e, nil
}

func (e *consoleEncoder) WithCallerEnabled() (dazl.Encoder, error) {
	e.caller = true
	return e, nil
}

func (e *consoleEncoder) WithStacktraceEnabled() (dazl.Encoder, error) {
	e.stacktrace = true
	return e, nil
}

type jsonEncoder struct {
	nameKey    string
	timestamp  bool
	caller     bool
	stacktrace bool
}

func (e *jsonEncoder) NewWriter(writer io.Writer) (dazl.Writer, error) {
	logger := zerolog.New(writer)
	if e.timestamp {
		logger = logger.With().Timestamp().Logger()
	}
	if e.caller {
		logger = logger.With().CallerWithSkipFrameCount(3).Logger()
	}
	if e.stacktrace {
		logger = logger.With().Stack().Logger()
	}
	return &Writer{
		logger:  logger,
		nameKey: e.nameKey,
	}, nil
}

func (e *jsonEncoder) WithMessageKey(key string) (dazl.Encoder, error) {
	zerolog.MessageFieldName = key
	return e, nil
}

func (e *jsonEncoder) WithNameEnabled() (dazl.Encoder, error) {
	return &jsonEncoder{
		nameKey: "logger",
	}, nil
}

func (e *jsonEncoder) WithNameKey(key string) (dazl.Encoder, error) {
	return &jsonEncoder{
		nameKey: key,
	}, nil
}

func (e *jsonEncoder) WithLevelEnabled() (dazl.Encoder, error) {
	return e, nil
}

func (e *jsonEncoder) WithLevelKey(key string) (dazl.Encoder, error) {
	zerolog.LevelFieldName = key
	return e, nil
}

func (e *jsonEncoder) WithLevelFormat(format dazl.LevelFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.LowerCaseLevelFormat:
		zerolog.LevelTraceValue = "trace"
		zerolog.LevelDebugValue = "debug"
		zerolog.LevelInfoValue = "info"
		zerolog.LevelWarnValue = "warn"
		zerolog.LevelErrorValue = "error"
		zerolog.LevelFatalValue = "fatal"
		zerolog.LevelPanicValue = "panic"
	case dazl.UpperCaseLevelFormat:
		zerolog.LevelTraceValue = "TRACE"
		zerolog.LevelDebugValue = "DEBUG"
		zerolog.LevelInfoValue = "INFO"
		zerolog.LevelWarnValue = "WARN"
		zerolog.LevelErrorValue = "ERROR"
		zerolog.LevelFatalValue = "FATAL"
		zerolog.LevelPanicValue = "PANIC"
	}
	return e, nil
}

func (e *jsonEncoder) WithTimestampEnabled() (dazl.Encoder, error) {
	e.timestamp = true
	return e, nil
}

func (e *jsonEncoder) WithTimestampKey(key string) (dazl.Encoder, error) {
	zerolog.TimestampFieldName = key
	return e, nil
}

func (e *jsonEncoder) WithTimestampFormat(format dazl.TimestampFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.UnixTimestampFormat:
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	case dazl.ISO8601TimestampFormat:
		zerolog.TimeFieldFormat = time.RFC3339
	default:
		return nil, fmt.Errorf("unsupoorted timestamp format %s", format)
	}
	return e, nil
}

func (e *jsonEncoder) WithCallerEnabled() (dazl.Encoder, error) {
	return e.WithCallerFormat(dazl.ShortCallerFormat)
}

func (e *jsonEncoder) WithCallerKey(key string) (dazl.Encoder, error) {
	e.caller = true
	zerolog.CallerFieldName = key
	return e, nil
}

func (e *jsonEncoder) WithCallerFormat(format dazl.CallerFormat) (dazl.Encoder, error) {
	switch format {
	case dazl.ShortCallerFormat:
		zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
			return filepath.Base(file) + ":" + strconv.Itoa(line)
		}
	case dazl.FullCallerFormat:
		zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
			return file + ":" + strconv.Itoa(line)
		}
	default:
		return nil, fmt.Errorf("unsupoorted caller format %s", format)
	}
	return e, nil
}

/*
func (e *jsonEncoder) WithStacktraceEnabled() (dazl.Encoder, error) {
	e.stacktrace = true
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	return e, nil
}

func (e *jsonEncoder) WithStacktraceKey(key string) (dazl.Encoder, error) {
	zerolog.ErrorStackFieldName = key
	return e, nil
}
*/

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

//var _ dazl.StacktraceEncoder = (*jsonEncoder)(nil)
//var _ dazl.StacktraceKeyEncoder = (*jsonEncoder)(nil)
