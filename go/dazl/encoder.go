// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

type Encoding string

const (
	ConsoleEncoding Encoding = "console"
	JSONEncoding    Encoding = "json"
)

type Encoder interface {
	NewWriter(writer io.Writer) (Writer, error)
}

type NameEncoder interface {
	WithNameEnabled() (Encoder, error)
}

type NameKeyEncoder interface {
	WithNameKey(key string) (Encoder, error)
}

type MessageKeyEncoder interface {
	WithMessageKey(key string) (Encoder, error)
}

type LevelEncoder interface {
	WithLevelEnabled() (Encoder, error)
}

type LevelKeyEncoder interface {
	WithLevelKey(key string) (Encoder, error)
}

type LevelFormattingEncoder interface {
	WithLevelFormat(format LevelFormat) (Encoder, error)
}

type TimestampEncoder interface {
	WithTimestampEnabled() (Encoder, error)
}

type TimestampKeyEncoder interface {
	WithTimestampKey(key string) (Encoder, error)
}

type TimestampFormattingEncoder interface {
	WithTimestampFormat(format TimestampFormat) (Encoder, error)
}

type CallerEncoder interface {
	WithCallerEnabled() (Encoder, error)
}

type CallerKeyEncoder interface {
	WithCallerKey(key string) (Encoder, error)
}

type CallerFormattingEncoder interface {
	WithCallerFormat(format CallerFormat) (Encoder, error)
}

type StacktraceEncoder interface {
	WithStacktraceEnabled() (Encoder, error)
}

type StacktraceKeyEncoder interface {
	WithStacktraceKey(key string) (Encoder, error)
}

func configureConsoleEncoder(config encoderConfig, encoder Encoder) (Encoder, error) {
	var err error
	if config.Fields.Name != nil {
		if nameEncoder, ok := encoder.(NameEncoder); ok {
			encoder, err = nameEncoder.WithNameEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("name is not an optional field for the console encoder")
		}
	}
	if config.Fields.Level != nil {
		if levelEncoder, ok := encoder.(LevelEncoder); ok {
			encoder, err = levelEncoder.WithLevelEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("level is not an optional field for the console encoder")
		}
		if config.Fields.Level.Format != nil {
			if levelFormattingEncoder, ok := encoder.(LevelFormattingEncoder); ok {
				encoder, err = levelFormattingEncoder.WithLevelFormat(*config.Fields.Level.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the console encoder does not support configuring level formats")
			}
		}
	}
	if config.Fields.Time != nil {
		if timestampEncoder, ok := encoder.(TimestampEncoder); ok {
			encoder, err = timestampEncoder.WithTimestampEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("time is not an optional field for the console encoder")
		}
		if config.Fields.Time.Format != nil {
			if timestampFormattingEncoder, ok := encoder.(TimestampFormattingEncoder); ok {
				encoder, err = timestampFormattingEncoder.WithTimestampFormat(*config.Fields.Time.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the console encoder does not support configuring timestamp formats")
			}
		}
	}
	if config.Fields.Caller != nil {
		if callerEncoder, ok := encoder.(CallerEncoder); ok {
			encoder, err = callerEncoder.WithCallerEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("caller is not an optional field for the console encoder")
		}
		if config.Fields.Caller.Format != nil {
			if callerFormattingEncoder, ok := encoder.(CallerFormattingEncoder); ok {
				encoder, err = callerFormattingEncoder.WithCallerFormat(*config.Fields.Caller.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the console encoder does not support configuring caller formats")
			}
		}
	}
	if config.Fields.Stacktrace != nil {
		if stacktraceEncoder, ok := encoder.(StacktraceEncoder); ok {
			encoder, err = stacktraceEncoder.WithStacktraceEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("stacktrace is not an optional field for the console encoder")
		}
	}
	return encoder, nil
}

func configureJSONEncoder(config encoderConfig, encoder Encoder) (Encoder, error) {
	var err error
	if config.Fields.Message != nil {
		if config.Fields.Message.Key != "" {
			if messageKeyEncoder, ok := encoder.(MessageKeyEncoder); ok {
				encoder, err = messageKeyEncoder.WithMessageKey(config.Fields.Message.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring message keys")
			}
		}
	}
	if config.Fields.Name != nil {
		if nameEncoder, ok := encoder.(NameEncoder); ok {
			encoder, err = nameEncoder.WithNameEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("name is not an optional field for the JSON encoder")
		}
		if config.Fields.Name.Key != "" {
			if nameKeyEncoder, ok := encoder.(NameKeyEncoder); ok {
				encoder, err = nameKeyEncoder.WithNameKey(config.Fields.Name.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring name keys")
			}
		}
	}
	if config.Fields.Level != nil {
		if levelEncoder, ok := encoder.(LevelEncoder); ok {
			encoder, err = levelEncoder.WithLevelEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("level is not an optional field for the JSON encoder")
		}
		if config.Fields.Level.Key != "" {
			if levelKeyEncoder, ok := encoder.(LevelKeyEncoder); ok {
				encoder, err = levelKeyEncoder.WithLevelKey(config.Fields.Level.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring level keys")
			}
		}
		if config.Fields.Level.Format != nil {
			if levelFormattingEncoder, ok := encoder.(LevelFormattingEncoder); ok {
				encoder, err = levelFormattingEncoder.WithLevelFormat(*config.Fields.Level.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring level formats")
			}
		}
	}
	if config.Fields.Time != nil {
		if timestampEncoder, ok := encoder.(TimestampEncoder); ok {
			encoder, err = timestampEncoder.WithTimestampEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("timestamp is not an optional field for the JSON encoder")
		}
		if config.Fields.Time.Key != "" {
			if timestampKeyEncoder, ok := encoder.(TimestampKeyEncoder); ok {
				encoder, err = timestampKeyEncoder.WithTimestampKey(config.Fields.Time.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring timestamp keys")
			}
		}
		if config.Fields.Time.Format != nil {
			if timestampFormattingEncoder, ok := encoder.(TimestampFormattingEncoder); ok {
				encoder, err = timestampFormattingEncoder.WithTimestampFormat(*config.Fields.Time.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring timestamp formats")
			}
		}
	}
	if config.Fields.Caller != nil {
		if callerEncoder, ok := encoder.(CallerEncoder); ok {
			encoder, err = callerEncoder.WithCallerEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("caller is not an optional field for the JSON encoder")
		}
		if config.Fields.Caller.Key != "" {
			if callerKeyEncoder, ok := encoder.(CallerKeyEncoder); ok {
				encoder, err = callerKeyEncoder.WithCallerKey(config.Fields.Caller.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring caller keys")
			}
		}
		if config.Fields.Caller.Format != nil {
			if callerFormattingEncoder, ok := encoder.(CallerFormattingEncoder); ok {
				encoder, err = callerFormattingEncoder.WithCallerFormat(*config.Fields.Caller.Format)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring caller formats")
			}
		}
	}
	if config.Fields.Stacktrace != nil {
		if stacktraceEncoder, ok := encoder.(StacktraceEncoder); ok {
			encoder, err = stacktraceEncoder.WithStacktraceEnabled()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("stacktrace is not an optional field for the JSON encoder")
		}
		if config.Fields.Stacktrace.Key != "" {
			if stacktraceKeyEncoder, ok := encoder.(StacktraceKeyEncoder); ok {
				encoder, err = stacktraceKeyEncoder.WithStacktraceKey(config.Fields.Stacktrace.Key)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("the JSON encoder does not support configuring stacktrace keys")
			}
		}
	}
	return encoder, nil
}

type encodersConfig struct {
	Console encoderConfig `json:"console" yaml:"console"`
	JSON    encoderConfig `json:"json" yaml:"json"`
}

type encoderConfig struct {
	Fields encoderFieldsConfig `json:"fields" yaml:"fields"`
}

type encoderFieldsConfig struct {
	Name       *nameEncoderConfig       `json:"name" yaml:"name"`
	Message    *messageEncoderConfig    `json:"message" yaml:"message"`
	Level      *levelEncoderConfig      `json:"level" yaml:"level"`
	Time       *timestampEncoderConfig  `json:"timestamp" yaml:"timestamp"`
	Caller     *callerEncoderConfig     `json:"caller" yaml:"caller"`
	Stacktrace *stacktraceEncoderConfig `json:"stacktrace" yaml:"stacktrace"`
}

func (c *encoderFieldsConfig) UnmarshalYAML(unmarshal func(any) error) error {
	object := make(map[string]any)
	if err := unmarshal(&object); err == nil {
		for name, value := range object {
			bytes, err := yaml.Marshal(value)
			if err != nil {
				return err
			}
			switch name {
			case "message":
				c.Message = &messageEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Message); err != nil {
					return err
				}
			case "name":
				c.Name = &nameEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Name); err != nil {
					return err
				}
			case "level":
				c.Level = &levelEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Level); err != nil {
					return err
				}
			case "timestamp":
				c.Time = &timestampEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Time); err != nil {
					return err
				}
			case "caller":
				c.Caller = &callerEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Caller); err != nil {
					return err
				}
			case "stacktrace":
				c.Stacktrace = &stacktraceEncoderConfig{}
				if err := yaml.Unmarshal(bytes, c.Stacktrace); err != nil {
					return err
				}
			}
		}
		return nil
	}

	var fields []encoderFieldSchema
	if err := unmarshal(&fields); err != nil {
		return err
	}
	for _, field := range fields {
		if field.Name != nil {
			c.Name = field.Name
		}
		if field.Message != nil {
			c.Message = field.Message
		}
		if field.Level != nil {
			c.Level = field.Level
		}
		if field.Time != nil {
			c.Time = field.Time
		}
		if field.Caller != nil {
			c.Caller = field.Caller
		}
		if field.Stacktrace != nil {
			c.Stacktrace = field.Stacktrace
		}
	}
	return nil
}

type encoderFieldSchema struct {
	Name       *nameEncoderConfig       `json:"name" yaml:"name"`
	Message    *messageEncoderConfig    `json:"message" yaml:"message"`
	Level      *levelEncoderConfig      `json:"level" yaml:"level"`
	Time       *timestampEncoderConfig  `json:"timestamp" yaml:"timestamp"`
	Caller     *callerEncoderConfig     `json:"caller" yaml:"caller"`
	Stacktrace *stacktraceEncoderConfig `json:"stacktrace" yaml:"stacktrace"`
}

func (c *encoderFieldSchema) UnmarshalYAML(unmarshal func(any) error) error {
	config := make(map[string]any)
	if err := unmarshal(&config); err != nil {
		var text string
		if err := unmarshal(&text); err != nil {
			return err
		}
		return c.UnmarshalText([]byte(text))
	}
	if len(config) > 1 {
		return fmt.Errorf("encoder fields must configure one encoder per list item")
	}
	for key, value := range config {
		text, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		name := fieldEncoderName(key)
		switch name {
		case nameFieldName:
			c.Name = &nameEncoderConfig{}
			return yaml.Unmarshal(text, c.Name)
		case messageFieldName:
			c.Message = &messageEncoderConfig{}
			return yaml.Unmarshal(text, c.Message)
		case levelFieldName:
			c.Level = &levelEncoderConfig{}
			return yaml.Unmarshal(text, c.Level)
		case timestampFieldName:
			c.Time = &timestampEncoderConfig{}
			return yaml.Unmarshal(text, c.Time)
		case callerFieldName:
			c.Caller = &callerEncoderConfig{}
			return yaml.Unmarshal(text, c.Caller)
		case stacktraceFieldName:
			c.Stacktrace = &stacktraceEncoderConfig{}
			return yaml.Unmarshal(text, c.Stacktrace)
		default:
			return fmt.Errorf("unknown field encoder '%s'", name)
		}
	}
	return nil
}

func (c *encoderFieldSchema) UnmarshalText(text []byte) error {
	name := fieldEncoderName(text)
	switch name {
	case nameFieldName:
		c.Name = &nameEncoderConfig{}
	case messageFieldName:
		c.Message = &messageEncoderConfig{}
	case levelFieldName:
		c.Level = &levelEncoderConfig{}
	case timestampFieldName:
		c.Time = &timestampEncoderConfig{}
	case callerFieldName:
		c.Caller = &callerEncoderConfig{}
	case stacktraceFieldName:
		c.Stacktrace = &stacktraceEncoderConfig{}
	default:
		return fmt.Errorf("unknown field encoder '%s'", name)
	}
	return nil
}

type fieldEncoderName string

const (
	nameFieldName       fieldEncoderName = "name"
	messageFieldName    fieldEncoderName = "message"
	levelFieldName      fieldEncoderName = "level"
	timestampFieldName  fieldEncoderName = "timestamp"
	callerFieldName     fieldEncoderName = "caller"
	stacktraceFieldName fieldEncoderName = "stacktrace"
)

type fieldEncoderConfig struct {
	Key string `json:"key" yaml:"key"`
}

type nameEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
}

type messageEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
}

type LevelFormat string

const (
	LowerCaseLevelFormat LevelFormat = "lowercase"
	UpperCaseLevelFormat LevelFormat = "uppercase"
)

func (f LevelFormat) String() string {
	return string(f)
}

func (f *LevelFormat) UnmarshalText(text []byte) error {
	name := string(text)
	switch strings.ToLower(name) {
	case LowerCaseLevelFormat.String():
		*f = LowerCaseLevelFormat
	case UpperCaseLevelFormat.String():
		*f = UpperCaseLevelFormat
	default:
		return fmt.Errorf("unknown level format '%s'", name)
	}
	return nil
}

type levelEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
	Format             *LevelFormat `json:"format" yaml:"format"`
}

type TimestampFormat string

const (
	ISO8601TimestampFormat TimestampFormat = "iso8601"
	UnixTimestampFormat    TimestampFormat = "unix"
)

func (f TimestampFormat) String() string {
	return string(f)
}

func (f *TimestampFormat) UnmarshalText(text []byte) error {
	name := string(text)
	switch strings.ToLower(name) {
	case ISO8601TimestampFormat.String():
		*f = ISO8601TimestampFormat
	case UnixTimestampFormat.String():
		*f = UnixTimestampFormat
	default:
		return fmt.Errorf("unknown timestamp format '%s'", name)
	}
	return nil
}

type timestampEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
	Format             *TimestampFormat `json:"format" yaml:"format"`
}

type CallerFormat string

const (
	ShortCallerFormat CallerFormat = "short"
	FullCallerFormat  CallerFormat = "full"
)

func (f CallerFormat) String() string {
	return string(f)
}

func (f *CallerFormat) UnmarshalText(text []byte) error {
	name := string(text)
	switch strings.ToLower(name) {
	case ShortCallerFormat.String():
		*f = ShortCallerFormat
	case FullCallerFormat.String():
		*f = FullCallerFormat
	default:
		return fmt.Errorf("unknown caller format '%s'", name)
	}
	return nil
}

type callerEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
	Format             *CallerFormat `json:"format" yaml:"format"`
}

type stacktraceEncoderConfig struct {
	fieldEncoderConfig `json:",inline" yaml:",inline"`
}
