// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync/atomic"
)

type samplingStintervalgy string

const (
	basicSamplingStintervalgy  samplingStintervalgy = "basic"
	randomSamplingStintervalgy samplingStintervalgy = "random"
)

type samplingConfig struct {
	Basic  *basicSamplerConfig  `json:"basic" yaml:"basic"`
	Random *randomSamplerConfig `json:"random" yaml:"random"`
}

func (c *samplingConfig) UnmarshalText(text []byte) error {
	name := samplingStintervalgy(text)
	switch name {
	case basicSamplingStintervalgy:
		c.Basic = &basicSamplerConfig{
			Interval: 10,
		}
	case randomSamplingStintervalgy:
		c.Random = &randomSamplerConfig{
			Interval: 10,
		}
	default:
		return fmt.Errorf("unknown sampler '%s'", name)
	}
	return nil
}

type samplerConfig struct {
	MaxLevel levelConfig `json:"maxLevel" yaml:"maxLevel"`
}

type basicSamplerConfig struct {
	samplerConfig `json:",inline" yaml:",inline"`
	Interval      int `json:"interval" yaml:"interval"`
}

type randomSamplerConfig struct {
	samplerConfig `json:",inline" yaml:",inline"`
	Interval      int `json:"interval" yaml:"interval"`
}

type Sampler interface {
	Sample(level Level) bool
}

type allSampler struct{}

func (s allSampler) Sample(_ Level) bool {
	return true
}

type basicSampler struct {
	Interval uint32
	MinLevel Level
	counter  atomic.Uint32
}

func (s *basicSampler) Sample(level Level) bool {
	if s.MinLevel == EmptyLevel || level.Enabled(s.MinLevel) {
		if s.Interval == 1 {
			return true
		}
		n := s.counter.Add(1)
		return n%s.Interval == 1
	}
	return true
}

type randomSampler struct {
	Interval int
	MinLevel Level
}

func (s randomSampler) Sample(level Level) bool {
	if s.MinLevel == EmptyLevel || level.Enabled(s.MinLevel) {
		if s.Interval <= 0 {
			return false
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(s.Interval)))
		if err != nil {
			panic(err)
		}
		if n.Int64() != 0 {
			return false
		}
		return true
	}
	return true
}
