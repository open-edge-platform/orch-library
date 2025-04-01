// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package dazl is a generated GoMock package.
package dazl

import (
	"io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockEncoder is a mock of Encoder interface.
type MockEncoder struct {
	ctrl     *gomock.Controller
	recorder *MockEncoderMockRecorder
}

// MockEncoderMockRecorder is the mock recorder for MockEncoder.
type MockEncoderMockRecorder struct {
	mock *MockEncoder
}

// NewMockEncoder creates a new mock instance.
func NewMockEncoder(ctrl *gomock.Controller) *MockEncoder {
	mock := &MockEncoder{ctrl: ctrl}
	mock.recorder = &MockEncoderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEncoder) EXPECT() *MockEncoderMockRecorder {
	return m.recorder
}

// NewWriter mocks base method.
func (m *MockEncoder) NewWriter(arg0 io.Writer) (Writer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewWriter", arg0)
	ret0, _ := ret[0].(Writer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewWriter indicates an expected call of NewWriter.
func (mr *MockEncoderMockRecorder) NewWriter(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewWriter", reflect.TypeOf((*MockEncoder)(nil).NewWriter), arg0)
}

// WithMessageKey mocks base method.
func (m *MockEncoder) WithMessageKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithMessageKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithMessageKey indicates an expected call of WithMessageKey.
func (mr *MockEncoderMockRecorder) WithMessageKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithMessageKey", reflect.TypeOf((*MockEncoder)(nil).WithMessageKey), arg0)
}

// WithNameEnabled mocks base method.
func (m *MockEncoder) WithNameEnabled() (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithNameEnabled")
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithNameEnabled indicates an expected call of WithNameEnabled.
func (mr *MockEncoderMockRecorder) WithNameEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithNameEnabled", reflect.TypeOf((*MockEncoder)(nil).WithNameEnabled))
}

// WithNameKey mocks base method.
func (m *MockEncoder) WithNameKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithNameKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithNameKey indicates an expected call of WithNameKey.
func (mr *MockEncoderMockRecorder) WithNameKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithNameKey", reflect.TypeOf((*MockEncoder)(nil).WithNameKey), arg0)
}

// WithLevelEnabled mocks base method.
func (m *MockEncoder) WithLevelEnabled() (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithLevelEnabled")
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithLevelEnabled indicates an expected call of WithLevelEnabled.
func (mr *MockEncoderMockRecorder) WithLevelEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithLevelEnabled", reflect.TypeOf((*MockEncoder)(nil).WithLevelEnabled))
}

// WithLevelKey mocks base method.
func (m *MockEncoder) WithLevelKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithLevelKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithLevelKey indicates an expected call of WithLevelKey.
func (mr *MockEncoderMockRecorder) WithLevelKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithLevelKey", reflect.TypeOf((*MockEncoder)(nil).WithLevelKey), arg0)
}

// WithLevelFormat mocks base method.
func (m *MockEncoder) WithLevelFormat(arg0 LevelFormat) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithLevelFormat", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithLevelFormat indicates an expected call of WithLevelFormat.
func (mr *MockEncoderMockRecorder) WithLevelFormat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithLevelFormat", reflect.TypeOf((*MockEncoder)(nil).WithLevelFormat), arg0)
}

// WithTimestampEnabled mocks base method.
func (m *MockEncoder) WithTimestampEnabled() (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTimestampEnabled")
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithTimestampEnabled indicates an expected call of WithTimestampEnabled.
func (mr *MockEncoderMockRecorder) WithTimestampEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTimestampEnabled", reflect.TypeOf((*MockEncoder)(nil).WithTimestampEnabled))
}

// WithTimestampKey mocks base method.
func (m *MockEncoder) WithTimestampKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTimestampKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithTimestampKey indicates an expected call of WithTimestampKey.
func (mr *MockEncoderMockRecorder) WithTimestampKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTimestampKey", reflect.TypeOf((*MockEncoder)(nil).WithTimestampKey), arg0)
}

// WithTimestampFormat mocks base method.
func (m *MockEncoder) WithTimestampFormat(arg0 TimestampFormat) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTimestampFormat", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithTimestampFormat indicates an expected call of WithTimestampFormat.
func (mr *MockEncoderMockRecorder) WithTimestampFormat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTimestampFormat", reflect.TypeOf((*MockEncoder)(nil).WithTimestampFormat), arg0)
}

// WithCallerEnabled mocks base method.
func (m *MockEncoder) WithCallerEnabled() (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithCallerEnabled")
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithCallerEnabled indicates an expected call of WithCallerEnabled.
func (mr *MockEncoderMockRecorder) WithCallerEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithCallerEnabled", reflect.TypeOf((*MockEncoder)(nil).WithCallerEnabled))
}

// WithCallerKey mocks base method.
func (m *MockEncoder) WithCallerKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithCallerKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithCallerKey indicates an expected call of WithCallerKey.
func (mr *MockEncoderMockRecorder) WithCallerKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithCallerKey", reflect.TypeOf((*MockEncoder)(nil).WithCallerKey), arg0)
}

// WithCallerFormat mocks base method.
func (m *MockEncoder) WithCallerFormat(arg0 CallerFormat) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithCallerFormat", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithCallerFormat indicates an expected call of WithCallerFormat.
func (mr *MockEncoderMockRecorder) WithCallerFormat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithCallerFormat", reflect.TypeOf((*MockEncoder)(nil).WithCallerFormat), arg0)
}

// WithStacktraceEnabled mocks base method.
func (m *MockEncoder) WithStacktraceEnabled() (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithStacktraceEnabled")
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithStacktraceEnabled indicates an expected call of WithStacktraceEnabled.
func (mr *MockEncoderMockRecorder) WithStacktraceEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithStacktraceEnabled", reflect.TypeOf((*MockEncoder)(nil).WithStacktraceEnabled))
}

// WithStacktraceKey mocks base method.
func (m *MockEncoder) WithStacktraceKey(arg0 string) (Encoder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithStacktraceKey", arg0)
	ret0, _ := ret[0].(Encoder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WithStacktraceKey indicates an expected call of WithStacktraceKey.
func (mr *MockEncoderMockRecorder) WithStacktraceKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithStacktraceKey", reflect.TypeOf((*MockEncoder)(nil).WithStacktraceKey), arg0)
}

// MockWriter is a mock of Writer interface.
type MockWriter struct {
	ctrl     *gomock.Controller
	recorder *MockWriterMockRecorder
}

// MockWriterMockRecorder is the mock recorder for MockWriter.
type MockWriterMockRecorder struct {
	mock *MockWriter
}

// NewMockWriter creates a new mock instance.
func NewMockWriter(ctrl *gomock.Controller) *MockWriter {
	mock := &MockWriter{ctrl: ctrl}
	mock.recorder = &MockWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWriter) EXPECT() *MockWriterMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockWriter) Debug(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockWriterMockRecorder) Debug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockWriter)(nil).Debug), arg0)
}

// Error mocks base method.
func (m *MockWriter) Error(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockWriterMockRecorder) Error(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockWriter)(nil).Error), arg0)
}

// Fatal mocks base method.
func (m *MockWriter) Fatal(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Fatal", arg0)
}

// Fatal indicates an expected call of Fatal.
func (mr *MockWriterMockRecorder) Fatal(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatal", reflect.TypeOf((*MockWriter)(nil).Fatal), arg0)
}

// Info mocks base method.
func (m *MockWriter) Info(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0)
}

// Info indicates an expected call of Info.
func (mr *MockWriterMockRecorder) Info(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockWriter)(nil).Info), arg0)
}

// Panic mocks base method.
func (m *MockWriter) Panic(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Panic", arg0)
}

// Panic indicates an expected call of Panic.
func (mr *MockWriterMockRecorder) Panic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Panic", reflect.TypeOf((*MockWriter)(nil).Panic), arg0)
}

// Warn mocks base method.
func (m *MockWriter) Warn(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0)
}

// Warn indicates an expected call of Warn.
func (mr *MockWriterMockRecorder) Warn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockWriter)(nil).Warn), arg0)
}

// WithName mocks base method.
func (m *MockWriter) WithName(arg0 string) Writer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithName", arg0)
	ret0, _ := ret[0].(Writer)
	return ret0
}

// WithName indicates an expected call of WithName.
func (mr *MockWriterMockRecorder) WithName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithName", reflect.TypeOf((*MockWriter)(nil).WithName), arg0)
}

// WithSkipCalls mocks base method.
func (m *MockWriter) WithSkipCalls(arg0 int) Writer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithSkipCalls", arg0)
	ret0, _ := ret[0].(Writer)
	return ret0
}

// WithSkipCalls indicates an expected call of WithSkipCalls.
func (mr *MockWriterMockRecorder) WithSkipCalls(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithSkipCalls", reflect.TypeOf((*MockWriter)(nil).WithSkipCalls), arg0)
}
