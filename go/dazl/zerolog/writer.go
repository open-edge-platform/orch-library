// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zerolog

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/rs/zerolog"
	"time"
)

// Writer is a dazl output implementation
type Writer struct {
	logger     zerolog.Logger
	nameKey    string
	name       string
	skipFrames int
}

func (w *Writer) Debug(msg string) {
	w.logger.Debug().Msg(msg)
}

func (w *Writer) Info(msg string) {
	w.logger.Info().Msg(msg)
}

func (w *Writer) Error(msg string) {
	w.logger.Error().Msg(msg)
}

func (w *Writer) Fatal(msg string) {
	w.logger.Fatal().Msg(msg)
}

func (w *Writer) Panic(msg string) {
	w.logger.Panic().Msg(msg)
}

func (w *Writer) Warn(msg string) {
	w.logger.Warn().Msg(msg)
}

func (w *Writer) withLogger(logger zerolog.Logger) dazl.Writer {
	return &Writer{
		logger:     logger,
		nameKey:    w.nameKey,
		name:       w.name,
		skipFrames: w.skipFrames,
	}
}

func (w *Writer) WithName(name string) dazl.Writer {
	if w.nameKey != "" {
		return &Writer{
			logger:     w.logger.With().Str(w.nameKey, name).Logger(),
			nameKey:    w.nameKey,
			name:       name,
			skipFrames: w.skipFrames,
		}
	}
	return &Writer{
		logger:     w.logger,
		nameKey:    name,
		name:       name,
		skipFrames: w.skipFrames,
	}
}

func (w *Writer) WithSkipCalls(calls int) dazl.Writer {
	return &Writer{
		logger:     w.logger.With().CallerWithSkipFrameCount(w.skipFrames + calls).Logger(),
		nameKey:    w.nameKey,
		name:       w.name,
		skipFrames: w.skipFrames + calls,
	}
}

func (w *Writer) WithErrorField(err error) dazl.Writer {
	return w.withLogger(w.logger.With().Err(err).Logger())
}

func (w *Writer) WithStringField(name string, value string) dazl.Writer {
	return w.withLogger(w.logger.With().Str(name, value).Logger())
}

func (w *Writer) WithBoolField(name string, value bool) dazl.Writer {
	return w.withLogger(w.logger.With().Bool(name, value).Logger())
}

func (w *Writer) WithIntField(name string, value int) dazl.Writer {
	return w.withLogger(w.logger.With().Int(name, value).Logger())
}

func (w *Writer) WithInt32Field(name string, value int32) dazl.Writer {
	return w.withLogger(w.logger.With().Int32(name, value).Logger())
}

func (w *Writer) WithInt64Field(name string, value int64) dazl.Writer {
	return w.withLogger(w.logger.With().Int64(name, value).Logger())
}

func (w *Writer) WithUintField(name string, value uint) dazl.Writer {
	return w.withLogger(w.logger.With().Uint(name, value).Logger())
}

func (w *Writer) WithUint32Field(name string, value uint32) dazl.Writer {
	return w.withLogger(w.logger.With().Uint32(name, value).Logger())
}

func (w *Writer) WithUint64Field(name string, value uint64) dazl.Writer {
	return w.withLogger(w.logger.With().Uint64(name, value).Logger())
}

func (w *Writer) WithFloat32Field(name string, value float32) dazl.Writer {
	return w.withLogger(w.logger.With().Float32(name, value).Logger())
}

func (w *Writer) WithFloat64Field(name string, value float64) dazl.Writer {
	return w.withLogger(w.logger.With().Float64(name, value).Logger())
}

func (w *Writer) WithTimeField(name string, value time.Time) dazl.Writer {
	return w.withLogger(w.logger.With().Time(name, value).Logger())
}

func (w *Writer) WithDurationField(name string, value time.Duration) dazl.Writer {
	return w.withLogger(w.logger.With().Dur(name, value).Logger())
}

func (w *Writer) WithBinaryField(name string, value []byte) dazl.Writer {
	return w.withLogger(w.logger.With().Hex(name, value).Logger())
}

func (w *Writer) WithBytesField(name string, value []byte) dazl.Writer {
	return w.withLogger(w.logger.With().Bytes(name, value).Logger())
}

func (w *Writer) WithStringSliceField(name string, values []string) dazl.Writer {
	return w.withLogger(w.logger.With().Strs(name, values).Logger())
}

func (w *Writer) WithBoolSliceField(name string, values []bool) dazl.Writer {
	return w.withLogger(w.logger.With().Bools(name, values).Logger())
}

func (w *Writer) WithIntSliceField(name string, values []int) dazl.Writer {
	return w.withLogger(w.logger.With().Ints(name, values).Logger())
}

func (w *Writer) WithInt32SliceField(name string, values []int32) dazl.Writer {
	return w.withLogger(w.logger.With().Ints32(name, values).Logger())
}

func (w *Writer) WithInt64SliceField(name string, values []int64) dazl.Writer {
	return w.withLogger(w.logger.With().Ints64(name, values).Logger())
}

func (w *Writer) WithUintSliceField(name string, values []uint) dazl.Writer {
	return w.withLogger(w.logger.With().Uints(name, values).Logger())
}

func (w *Writer) WithUint32SliceField(name string, values []uint32) dazl.Writer {
	return w.withLogger(w.logger.With().Uints32(name, values).Logger())
}

func (w *Writer) WithUint64SliceField(name string, values []uint64) dazl.Writer {
	return w.withLogger(w.logger.With().Uints64(name, values).Logger())
}

func (w *Writer) WithFloat32SliceField(name string, values []float32) dazl.Writer {
	return w.withLogger(w.logger.With().Floats32(name, values).Logger())
}

func (w *Writer) WithFloat64SliceField(name string, values []float64) dazl.Writer {
	return w.withLogger(w.logger.With().Floats64(name, values).Logger())
}

func (w *Writer) WithTimeSliceField(name string, values []time.Time) dazl.Writer {
	return w.withLogger(w.logger.With().Times(name, values).Logger())
}

func (w *Writer) WithDurationSliceField(name string, values []time.Duration) dazl.Writer {
	return w.withLogger(w.logger.With().Durs(name, values).Logger())
}

var _ dazl.Writer = (*Writer)(nil)
var _ dazl.FieldWriter = (*Writer)(nil)
