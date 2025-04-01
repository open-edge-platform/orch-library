// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type outputsConfig struct {
	Outputs map[string]outputSchema
}

func (c *outputsConfig) UnmarshalYAML(unmarshal func(any) error) error {
	config := make(map[string]outputSchema)
	if err := unmarshal(&config); err != nil {
		var config []outputConfig
		if err := unmarshal(&config); err != nil {
			return err
		}
		c.Outputs = make(map[string]outputSchema)
		for _, output := range config {
			c.Outputs[output.Writer] = outputSchema{
				Level:  output.Level,
				Sample: output.Sample,
			}
		}
		return nil
	}
	c.Outputs = config
	return nil
}

type outputConfig struct {
	Writer string         `json:"writer" yaml:"writer"`
	Level  levelConfig    `json:"level" yaml:"level"`
	Sample samplingConfig `json:"sample" yaml:"sample"`
}

func (c *outputConfig) UnmarshalYAML(unmarshal func(any) error) error {
	config := make(map[string]any)
	if err := unmarshal(&config); err != nil {
		var text string
		if err := unmarshal(&text); err != nil {
			return err
		}
		return c.UnmarshalText([]byte(text))
	}
	if len(config) > 1 {
		return fmt.Errorf("logger outputs must configure one writer per list item")
	}
	for key, value := range config {
		text, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		var schema outputSchema
		if err := yaml.Unmarshal(text, &schema); err != nil {
			return err
		}
		c.Writer = key
		c.Level = schema.Level
		c.Sample = schema.Sample
	}
	return nil
}

func (c *outputConfig) UnmarshalText(text []byte) error {
	c.Writer = string(text)
	return nil
}

type outputSchema struct {
	Level  levelConfig    `json:"level" yaml:"level"`
	Sample samplingConfig `json:"sample" yaml:"sample"`
}

func newOutput(writer Writer, level Level, sampler Sampler) *dazlOutput {
	return &dazlOutput{
		writer:  writer,
		level:   level,
		sampler: sampler,
	}
}

// dazlOutput is a dazl output implementation
type dazlOutput struct {
	writer  Writer
	level   Level
	sampler Sampler
}

func (o *dazlOutput) WithWriter(writer Writer) *dazlOutput {
	return &dazlOutput{
		writer:  writer,
		level:   o.level,
		sampler: o.sampler,
	}
}

func (o *dazlOutput) Level() Level {
	return o.level
}

func (o *dazlOutput) WithLevel(level Level) *dazlOutput {
	return &dazlOutput{
		writer:  o.writer,
		level:   level,
		sampler: o.sampler,
	}
}

func (o *dazlOutput) WithSampler(sampler Sampler) *dazlOutput {
	return &dazlOutput{
		writer:  o.writer,
		level:   o.level,
		sampler: sampler,
	}
}

func (o *dazlOutput) Debug(msg string) {
	if o.level.Enabled(DebugLevel) && o.sampler.Sample(DebugLevel) {
		o.writer.Debug(msg)
	}
}

func (o *dazlOutput) Info(msg string) {
	if o.level.Enabled(InfoLevel) && o.sampler.Sample(InfoLevel) {
		o.writer.Info(msg)
	}
}

func (o *dazlOutput) Warn(msg string) {
	if o.level.Enabled(WarnLevel) && o.sampler.Sample(WarnLevel) {
		o.writer.Warn(msg)
	}
}

func (o *dazlOutput) Error(msg string) {
	if o.level.Enabled(ErrorLevel) && o.sampler.Sample(ErrorLevel) {
		o.writer.Error(msg)
	}
}

func (o *dazlOutput) Fatal(msg string) {
	if o.level.Enabled(FatalLevel) && o.sampler.Sample(FatalLevel) {
		o.writer.Fatal(msg)
	}
}

func (o *dazlOutput) Panic(msg string) {
	if o.level.Enabled(PanicLevel) && o.sampler.Sample(PanicLevel) {
		o.writer.Panic(msg)
	}
}
