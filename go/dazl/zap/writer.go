// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package zap

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

func newWriter(writer io.Writer, encoder zapcore.Encoder, config zap.Config) (dazl.Writer, error) {
	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.WrapCore(func(zapcore.Core) zapcore.Core {
			return zapcore.NewCore(encoder, &writeSyncer{writer}, zap.DebugLevel)
		}))
	if err != nil {
		return nil, err
	}

	return &Writer{
		root:   logger,
		logger: logger,
	}, nil
}

// Writer is a dazl output implementation
type Writer struct {
	root   *zap.Logger
	logger *zap.Logger
}

func (w *Writer) WithName(name string) dazl.Writer {
	return &Writer{
		root:   w.root,
		logger: w.root.Named(name),
	}
}

func (w *Writer) withField(field zap.Field) dazl.Writer {
	return &Writer{
		root:   w.root,
		logger: w.logger.With(field),
	}
}

func (w *Writer) WithStringField(name string, value string) dazl.Writer {
	return w.withField(zap.String(name, value))
}

func (w *Writer) WithBoolField(name string, value bool) dazl.Writer {
	return w.withField(zap.Bool(name, value))
}

func (w *Writer) WithIntField(name string, value int) dazl.Writer {
	return w.withField(zap.Int(name, value))
}

func (w *Writer) WithInt32Field(name string, value int32) dazl.Writer {
	return w.withField(zap.Int32(name, value))
}

func (w *Writer) WithInt64Field(name string, value int64) dazl.Writer {
	return w.withField(zap.Int64(name, value))
}

func (w *Writer) WithUintField(name string, value uint) dazl.Writer {
	return w.withField(zap.Uint(name, value))
}

func (w *Writer) WithUint32Field(name string, value uint32) dazl.Writer {
	return w.withField(zap.Uint32(name, value))
}

func (w *Writer) WithUint64Field(name string, value uint64) dazl.Writer {
	return w.withField(zap.Uint64(name, value))
}

func (w *Writer) WithFloat32Field(name string, value float32) dazl.Writer {
	return w.withField(zap.Float32(name, value))
}

func (w *Writer) WithFloat64Field(name string, value float64) dazl.Writer {
	return w.withField(zap.Float64(name, value))
}

func (w *Writer) WithTimeField(name string, value time.Time) dazl.Writer {
	return w.withField(zap.Time(name, value))
}

func (w *Writer) WithDurationField(name string, value time.Duration) dazl.Writer {
	return w.withField(zap.Duration(name, value))
}

func (w *Writer) WithBinaryField(name string, value []byte) dazl.Writer {
	return w.withField(zap.Binary(name, value))
}

func (w *Writer) WithBytesField(name string, value []byte) dazl.Writer {
	return w.withField(zap.ByteString(name, value))
}

func (w *Writer) WithStringSliceField(name string, values []string) dazl.Writer {
	return w.withField(zap.Strings(name, values))
}

func (w *Writer) WithBoolSliceField(name string, values []bool) dazl.Writer {
	return w.withField(zap.Bools(name, values))
}

func (w *Writer) WithIntSliceField(name string, values []int) dazl.Writer {
	return w.withField(zap.Ints(name, values))
}

func (w *Writer) WithInt32SliceField(name string, values []int32) dazl.Writer {
	return w.withField(zap.Int32s(name, values))
}

func (w *Writer) WithInt64SliceField(name string, values []int64) dazl.Writer {
	return w.withField(zap.Int64s(name, values))
}

func (w *Writer) WithUintSliceField(name string, values []uint) dazl.Writer {
	return w.withField(zap.Uints(name, values))
}

func (w *Writer) WithUint32SliceField(name string, values []uint32) dazl.Writer {
	return w.withField(zap.Uint32s(name, values))
}

func (w *Writer) WithUint64SliceField(name string, values []uint64) dazl.Writer {
	return w.withField(zap.Uint64s(name, values))
}

func (w *Writer) WithFloat32SliceField(name string, values []float32) dazl.Writer {
	return w.withField(zap.Float32s(name, values))
}

func (w *Writer) WithFloat64SliceField(name string, values []float64) dazl.Writer {
	return w.withField(zap.Float64s(name, values))
}

func (w *Writer) WithTimeSliceField(name string, values []time.Time) dazl.Writer {
	return w.withField(zap.Times(name, values))
}

func (w *Writer) WithDurationSliceField(name string, values []time.Duration) dazl.Writer {
	return w.withField(zap.Durations(name, values))
}

func (w *Writer) WithSkipCalls(calls int) dazl.Writer {
	return &Writer{
		root:   w.root.WithOptions(zap.AddCallerSkip(calls)),
		logger: w.logger.WithOptions(zap.AddCallerSkip(calls)),
	}
}

func (w *Writer) Debug(msg string) {
	w.logger.Debug(msg)
}

func (w *Writer) Info(msg string) {
	w.logger.Info(msg)
}

func (w *Writer) Error(msg string) {
	w.logger.Error(msg)
}

func (w *Writer) Fatal(msg string) {
	w.logger.Fatal(msg)
}

func (w *Writer) Panic(msg string) {
	w.logger.Panic(msg)
}

func (w *Writer) Warn(msg string) {
	w.logger.Warn(msg)
}

func (w *Writer) Sync() error {
	return w.logger.Sync()
}

var _ dazl.Writer = (*Writer)(nil)
var _ dazl.FieldWriter = (*Writer)(nil)

type writeSyncer struct {
	io.Writer
}

func (w *writeSyncer) Sync() error {
	return nil
}
