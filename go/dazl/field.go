// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Field is a structured logger field
type Field func(writer Writer) (Writer, error)

// Error creates a field for an error
func Error(err error) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(ErrorFieldWriter); ok {
			return fieldWriter.WithErrorField(err), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField("error", err.Error()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Stringer creates a named field for a type implementing Stringer
func Stringer(name string, value fmt.Stringer) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(StringerFieldWriter); ok {
			return fieldWriter.WithStringerField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, value.String()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// String creates a named string field
func String(name string, value string) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, value), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Strings creates a named string slice field
func Strings(name string, value []string) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(StringSliceFieldWriter); ok {
			return fieldWriter.WithStringSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Int creates a named int field
func Int(name string, value int) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.Itoa(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Ints creates a named int slice field
func Ints(name string, value []int) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(IntSliceFieldWriter); ok {
			return fieldWriter.WithIntSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Int32 creates a named int32 field
func Int32(name string, value int32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Int32FieldWriter); ok {
			return fieldWriter.WithInt32Field(name, value), nil
		}
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, int(value)), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Int32s creates a named int32 slice field
func Int32s(name string, value []int32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Int32SliceFieldWriter); ok {
			return fieldWriter.WithInt32SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Int64 creates a named int64 field
func Int64(name string, value int64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Int64FieldWriter); ok {
			return fieldWriter.WithInt64Field(name, value), nil
		}
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, int(value)), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.Itoa(int(value))), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Int64s creates a named int64 slice field
func Int64s(name string, value []int64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Int64SliceFieldWriter); ok {
			return fieldWriter.WithInt64SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uint creates a named uint field
func Uint(name string, value uint) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(UintFieldWriter); ok {
			return fieldWriter.WithUintField(name, value), nil
		}
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, int(value)), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.Itoa(int(value))), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uints creates a named uint slice field
func Uints(name string, value []uint) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(UintSliceFieldWriter); ok {
			return fieldWriter.WithUintSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uint32 creates a named uint32 field
func Uint32(name string, value uint32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Uint32FieldWriter); ok {
			return fieldWriter.WithUint32Field(name, value), nil
		}
		if fieldWriter, ok := writer.(UintFieldWriter); ok {
			return fieldWriter.WithUintField(name, uint(value)), nil
		}
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, int(value)), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.Itoa(int(value))), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uint32s creates a named uint32 slice field
func Uint32s(name string, value []uint32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Uint32SliceFieldWriter); ok {
			return fieldWriter.WithUint32SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uint64 creates a named uint64 field
func Uint64(name string, value uint64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Uint64FieldWriter); ok {
			return fieldWriter.WithUint64Field(name, value), nil
		}
		if fieldWriter, ok := writer.(UintFieldWriter); ok {
			return fieldWriter.WithUintField(name, uint(value)), nil
		}
		if fieldWriter, ok := writer.(IntFieldWriter); ok {
			return fieldWriter.WithIntField(name, int(value)), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.Itoa(int(value))), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Uint64s creates a named uint64 slice field
func Uint64s(name string, value []uint64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Uint64SliceFieldWriter); ok {
			return fieldWriter.WithUint64SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Float32 creates a named float32 field
func Float32(name string, value float32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Float32FieldWriter); ok {
			return fieldWriter.WithFloat32Field(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.FormatFloat(float64(value), 'f', -1, 32)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Float32s creates a named float32 slice field
func Float32s(name string, value []float32) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Float32SliceFieldWriter); ok {
			return fieldWriter.WithFloat32SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Float64 creates a named float64 field
func Float64(name string, value float64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Float64FieldWriter); ok {
			return fieldWriter.WithFloat64Field(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.FormatFloat(value, 'f', -1, 64)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Float64s creates a named float64 slice field
func Float64s(name string, value []float64) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(Float64SliceFieldWriter); ok {
			return fieldWriter.WithFloat64SliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Bool creates a named bool field
func Bool(name string, value bool) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(BoolFieldWriter); ok {
			return fieldWriter.WithBoolField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, strconv.FormatBool(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Bools creates a named bool slice field
func Bools(name string, value []bool) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(BoolSliceFieldWriter); ok {
			return fieldWriter.WithBoolSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Time creates a named Time field
func Time(name string, value time.Time) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(TimeFieldWriter); ok {
			return fieldWriter.WithTimeField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, value.String()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Times creates a named Time slice field
func Times(name string, value []time.Time) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(TimeSliceFieldWriter); ok {
			return fieldWriter.WithTimeSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Duration creates a named Duration field
func Duration(name string, value time.Duration) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(DurationFieldWriter); ok {
			return fieldWriter.WithDurationField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, value.String()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Durations creates a named Duration slice field
func Durations(name string, value []time.Duration) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(DurationSliceFieldWriter); ok {
			return fieldWriter.WithDurationSliceField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, fmt.Sprint(value)), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Binary creates a named binary field
func Binary(name string, value []byte) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(BinaryFieldWriter); ok {
			return fieldWriter.WithBinaryField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, bytes.NewBuffer(value).String()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}

// Bytes creates a named byte string field
func Bytes(name string, value []byte) Field {
	return func(writer Writer) (Writer, error) {
		if fieldWriter, ok := writer.(BytesFieldWriter); ok {
			return fieldWriter.WithBytesField(name, value), nil
		}
		if fieldWriter, ok := writer.(StringFieldWriter); ok {
			return fieldWriter.WithStringField(name, bytes.NewBuffer(value).String()), nil
		}
		return nil, fmt.Errorf("field type is not supported by the configured logging framework")
	}
}
