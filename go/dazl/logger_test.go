// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dazl

import (
	"bytes"
	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io"
	"sort"
	"strings"
	"testing"
)

func TestLoggerNames(t *testing.T) {
	assert.Equal(t, "", root.Name())
	assert.Equal(t, "foo", GetLogger("foo").Name())
	assert.Equal(t, "foo/bar", GetLogger("foo/bar").Name())
	assert.Equal(t, "github.com/open-edge-platform/orch-library/go/dazl", GetLogger().Name())
	assert.Equal(t, "github.com/open-edge-platform/orch-library/go/dazl", GetPackageLogger().Name())
}

func TestLoggerLevels(t *testing.T) {
	assert.Equal(t, EmptyLevel, GetRootLogger().Level())
	assert.Equal(t, EmptyLevel, GetLogger("foo").Level())
	assert.Equal(t, EmptyLevel, GetLogger("foo/bar/baz").Level())
	GetRootLogger().SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, GetRootLogger().Level())
	assert.Equal(t, DebugLevel, GetLogger("foo").Level())
	assert.Equal(t, DebugLevel, GetLogger("foo/bar/baz").Level())
	GetLogger("foo").SetLevel(InfoLevel)
	assert.Equal(t, DebugLevel, GetRootLogger().Level())
	assert.Equal(t, InfoLevel, GetLogger("foo").Level())
	assert.Equal(t, InfoLevel, GetLogger("foo/bar/baz").Level())
	GetRootLogger().SetLevel(WarnLevel)
	assert.Equal(t, WarnLevel, GetRootLogger().Level())
	assert.Equal(t, InfoLevel, GetLogger("foo").Level())
	assert.Equal(t, InfoLevel, GetLogger("foo/bar/baz").Level())
}

const testLoggerConfigArray = `
level: debug
sample:
  random:
    interval: 10
    level: debug
outputs:
  - stdout:
      level: info
  - stderr:
      level: error
  - file
`

func TestLoggerConfigArray(t *testing.T) {
	var config loggerConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testLoggerConfigArray), &config))

	assert.Equal(t, DebugLevel, config.Level.Level())
	assert.Len(t, config.Outputs.Outputs, 3)
	assert.Equal(t, InfoLevel, config.Outputs.Outputs["stdout"].Level.Level())
	assert.Equal(t, ErrorLevel, config.Outputs.Outputs["stderr"].Level.Level())
	assert.Equal(t, EmptyLevel, config.Outputs.Outputs["file"].Level.Level())
}

const testLoggerConfigObject = `
level: debug
sample:
  random:
    interval: 10
    level: debug
outputs:
  stdout:
    level: info
  stderr:
    level: error
  file:
`

func TestLoggerConfigObject(t *testing.T) {
	var config loggerConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testLoggerConfigObject), &config))

	assert.Equal(t, DebugLevel, config.Level.Level())
	assert.Len(t, config.Outputs.Outputs, 3)
	assert.Equal(t, InfoLevel, config.Outputs.Outputs["stdout"].Level.Level())
	assert.Equal(t, ErrorLevel, config.Outputs.Outputs["stderr"].Level.Level())
	assert.Equal(t, EmptyLevel, config.Outputs.Outputs["file"].Level.Level())
}

const testConfig = `
encoders:
  console:
    fields:
      - name
      - message
      - level:
          format: uppercase
      - timestamp:
          format: iso8601
      - caller:
          format: short
  json:
    fields:
      - name:
          key: logger
      - message
      - level:
          format: lowercase
      - timestamp:
          key: timestamp
      - caller
      - stacktrace:
          key: trace

writers:
  stdout:
    encoder: console
  file:
    encoder: json
    path: test.log

rootLogger:
  level: info
  outputs:
    - stdout

loggers:
  test/level:
    level: warn
  test/sample:
    sample:
      basic: 
        interval: 2
        maxLevel: warn
  test/sample/outputs:
    outputs:
      - file
  test/outputs:
    outputs:
      - file
  test/outputs/level:
    outputs:
      stdout:
        level: warn
  test/outputs/sample:
    outputs:
      stdout:
        sample:
          basic:
            interval: 2
            maxLevel: warn
`

func TestLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	console := NewMockEncoder(ctrl)
	json := NewMockEncoder(ctrl)

	stdout := NewMockWriter(ctrl)
	stdout.EXPECT().WithSkipCalls(gomock.Eq(2)).Return(stdout)
	file := NewMockWriter(ctrl)
	file.EXPECT().WithSkipCalls(gomock.Eq(2)).Return(file)

	console.EXPECT().NewWriter(gomock.Any()).Return(stdout, nil)
	json.EXPECT().NewWriter(gomock.Any()).Return(file, nil)

	console.EXPECT().WithNameEnabled().Return(console, nil)
	console.EXPECT().WithLevelEnabled().Return(console, nil)
	console.EXPECT().WithLevelFormat(gomock.Eq(UpperCaseLevelFormat)).Return(console, nil)
	console.EXPECT().WithTimestampEnabled().Return(console, nil)
	console.EXPECT().WithTimestampFormat(gomock.Eq(ISO8601TimestampFormat)).Return(console, nil)
	console.EXPECT().WithCallerEnabled().Return(console, nil)
	console.EXPECT().WithCallerFormat(gomock.Eq(ShortCallerFormat)).Return(console, nil)

	json.EXPECT().WithNameEnabled().Return(json, nil)
	json.EXPECT().WithNameKey(gomock.Eq("logger")).Return(json, nil)
	json.EXPECT().WithLevelEnabled().Return(json, nil)
	json.EXPECT().WithLevelFormat(gomock.Eq(LowerCaseLevelFormat)).Return(json, nil)
	json.EXPECT().WithTimestampEnabled().Return(json, nil)
	json.EXPECT().WithTimestampKey(gomock.Eq("timestamp")).Return(json, nil)
	json.EXPECT().WithCallerEnabled().Return(json, nil)
	json.EXPECT().WithStacktraceEnabled().Return(json, nil)
	json.EXPECT().WithStacktraceKey(gomock.Eq("trace")).Return(json, nil)

	framework := &testFramework{
		console: console,
		json:    json,
	}

	var config loggingConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testConfig), &config))
	assert.NoError(t, configure(framework, config, func(_ string) (io.Writer, error) {
		return &bytes.Buffer{}, nil
	}))

	stdout.EXPECT().WithName(gomock.Eq("test")).Return(stdout)
	log := GetLogger("test")

	log.Debug("debug")
	stdout.EXPECT().Info(gomock.Eq("info"))
	log.Info("info")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")

	stdout.EXPECT().WithName(gomock.Eq("test/level")).Return(stdout)
	log = GetLogger("test/level")

	log.Debug("debug")
	log.Info("info")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")

	stdout.EXPECT().WithName(gomock.Eq("test/sample")).Return(stdout)
	log = GetLogger("test/sample")

	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	log.Warn("warn")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	log.Warn("warn")

	stdout.EXPECT().WithName(gomock.Eq("test/sample/outputs")).Return(stdout)
	file.EXPECT().WithName(gomock.Eq("test/sample/outputs")).Return(file)
	log = GetLogger("test/sample/outputs")

	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	log.Warn("warn")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	log.Warn("warn")

	stdout.EXPECT().WithName(gomock.Eq("test/outputs")).Return(stdout)
	file.EXPECT().WithName(gomock.Eq("test/outputs")).Return(file)
	log = GetLogger("test/outputs")

	log.Debug("debug")
	stdout.EXPECT().Info(gomock.Eq("info"))
	file.EXPECT().Info(gomock.Eq("info"))
	log.Info("info")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")

	stdout.EXPECT().WithName(gomock.Eq("test/outputs/level")).Return(stdout)
	file.EXPECT().WithName(gomock.Eq("test/outputs/level")).Return(file)
	log = GetLogger("test/outputs/level")

	log.Debug("debug")
	file.EXPECT().Info(gomock.Eq("info"))
	log.Info("info")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")

	stdout.EXPECT().WithName(gomock.Eq("test/outputs/sample")).Return(stdout)
	file.EXPECT().WithName(gomock.Eq("test/outputs/sample")).Return(file)
	log = GetLogger("test/outputs/sample")

	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Error(gomock.Eq("error"))
	file.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	file.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
}

const testMethodsConfig = `
encoders:
  json:
    fields:
      - name:
          key: logger
      - message
      - level:
          format: lowercase

writers:
  stdout:
    encoder: json

rootLogger:
  level: debug
  outputs:
    - stdout
`

func TestLoggerMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	json := NewMockEncoder(ctrl)

	stdout := NewMockWriter(ctrl)
	stdout.EXPECT().WithSkipCalls(gomock.Eq(2)).Return(stdout)

	json.EXPECT().NewWriter(gomock.Any()).Return(stdout, nil)

	json.EXPECT().WithNameEnabled().Return(json, nil)
	json.EXPECT().WithNameKey(gomock.Eq("logger")).Return(json, nil)
	json.EXPECT().WithLevelEnabled().Return(json, nil)
	json.EXPECT().WithLevelFormat(gomock.Eq(LowerCaseLevelFormat)).Return(json, nil)

	framework := &testFramework{
		json: json,
	}

	var config loggingConfig
	assert.NoError(t, yaml.Unmarshal([]byte(testMethodsConfig), &config))
	assert.NoError(t, configure(framework, config, func(_ string) (io.Writer, error) {
		return &bytes.Buffer{}, nil
	}))

	stdout.EXPECT().WithName(gomock.Eq("test")).Return(stdout)
	stdout.EXPECT().WithSkipCalls(gomock.Eq(1)).Return(stdout).AnyTimes()
	log := GetLogger("test")

	stdout.EXPECT().Debug(gomock.Eq("debug"))
	log.Debug("debug")
	stdout.EXPECT().Debug(gomock.Eq("debug"))
	log.Debugf("debug")
	stdout.EXPECT().Debug(gomock.Eq("debug"))
	log.Debugw("debug")

	stdout.EXPECT().Info(gomock.Eq("info"))
	log.Info("info")
	stdout.EXPECT().Info(gomock.Eq("info"))
	log.Infof("info")
	stdout.EXPECT().Info(gomock.Eq("info"))
	log.Infow("info")

	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warn("warn")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warnf("warn")
	stdout.EXPECT().Warn(gomock.Eq("warn"))
	log.Warnw("warn")

	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Error("error")
	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Errorf("error")
	stdout.EXPECT().Error(gomock.Eq("error"))
	log.Errorw("error")
}

type testFramework struct {
	console Encoder
	json    Encoder
}

func (f *testFramework) Name() string {
	return "test"
}

func (f *testFramework) ConsoleEncoder() Encoder {
	return f.console
}

func (f *testFramework) JSONEncoder() Encoder {
	return f.json
}

var loggerLevels = map[int]Level{
	0: EmptyLevel,
	1: DebugLevel,
	2: InfoLevel,
	3: WarnLevel,
	4: ErrorLevel,
	5: PanicLevel,
	6: FatalLevel,
}

var levelFormats = map[int]LevelFormat{
	0: LowerCaseLevelFormat,
	1: UpperCaseLevelFormat,
}

var timestampFormats = map[int]TimestampFormat{
	0: UnixTimestampFormat,
	1: ISO8601TimestampFormat,
}

var callerFormats = map[int]CallerFormat{
	0: ShortCallerFormat,
	1: FullCallerFormat,
}

var encodings = map[int]Encoding{
	0: JSONEncoding,
	1: ConsoleEncoding,
}

const maxNumFileWriters = 10
const maxNumLoggers = 100

const alphabetChars = "abcdefghijklmnopqrstuvwxyz"
const loggerNameChars = alphabetChars + "/"

func FuzzLogger(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		config, err := getLoggingConfig(data)
		if err != nil {
			return
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		console := NewMockEncoder(ctrl)
		if config.Encoders.Console.Fields.Name != nil {
			console.EXPECT().WithNameEnabled().Return(console, nil).AnyTimes()
		}
		if config.Encoders.Console.Fields.Level != nil {
			console.EXPECT().WithLevelEnabled().Return(console, nil).AnyTimes()
			if config.Encoders.Console.Fields.Level.Format != nil {
				console.EXPECT().WithLevelFormat(gomock.Eq(*config.Encoders.Console.Fields.Level.Format)).Return(console, nil).AnyTimes()
			}
		}
		if config.Encoders.Console.Fields.Time != nil {
			console.EXPECT().WithTimestampEnabled().Return(console, nil).AnyTimes()
			if config.Encoders.Console.Fields.Time.Format != nil {
				console.EXPECT().WithTimestampFormat(gomock.Eq(*config.Encoders.Console.Fields.Time.Format)).Return(console, nil).AnyTimes()
			}
		}
		if config.Encoders.Console.Fields.Caller != nil {
			console.EXPECT().WithCallerEnabled().Return(console, nil).AnyTimes()
			if config.Encoders.Console.Fields.Caller.Format != nil {
				console.EXPECT().WithCallerFormat(gomock.Eq(*config.Encoders.Console.Fields.Caller.Format)).Return(console, nil).AnyTimes()
			}
		}
		if config.Encoders.Console.Fields.Stacktrace != nil {
			console.EXPECT().WithStacktraceEnabled().Return(console, nil).AnyTimes()
		}

		json := NewMockEncoder(ctrl)
		if config.Encoders.JSON.Fields.Message != nil && config.Encoders.JSON.Fields.Message.Key != "" {
			json.EXPECT().WithMessageKey(gomock.Eq(config.Encoders.JSON.Fields.Message.Key)).Return(json, nil).AnyTimes()
		}
		if config.Encoders.JSON.Fields.Name != nil {
			json.EXPECT().WithNameEnabled().Return(json, nil).AnyTimes()
			if config.Encoders.JSON.Fields.Name.Key != "" {
				json.EXPECT().WithNameKey(gomock.Eq(config.Encoders.JSON.Fields.Name.Key)).Return(json, nil).AnyTimes()
			}
		}
		if config.Encoders.JSON.Fields.Level != nil {
			json.EXPECT().WithLevelEnabled().Return(json, nil).AnyTimes()
			if config.Encoders.JSON.Fields.Level.Key != "" {
				json.EXPECT().WithLevelKey(gomock.Eq(config.Encoders.JSON.Fields.Level.Key)).Return(json, nil).AnyTimes()
			}
			if config.Encoders.JSON.Fields.Level.Format != nil {
				json.EXPECT().WithLevelFormat(gomock.Eq(*config.Encoders.JSON.Fields.Level.Format)).Return(json, nil).AnyTimes()
			}
		}
		if config.Encoders.JSON.Fields.Time != nil {
			json.EXPECT().WithTimestampEnabled().Return(json, nil).AnyTimes()
			if config.Encoders.JSON.Fields.Time.Key != "" {
				json.EXPECT().WithTimestampKey(gomock.Eq(config.Encoders.JSON.Fields.Time.Key)).Return(json, nil).AnyTimes()
			}
			if config.Encoders.JSON.Fields.Time.Format != nil {
				json.EXPECT().WithTimestampFormat(gomock.Eq(*config.Encoders.JSON.Fields.Time.Format)).Return(json, nil).AnyTimes()
			}
		}
		if config.Encoders.JSON.Fields.Caller != nil {
			json.EXPECT().WithCallerEnabled().Return(json, nil).AnyTimes()
			if config.Encoders.JSON.Fields.Caller.Key != "" {
				json.EXPECT().WithCallerKey(gomock.Eq(config.Encoders.JSON.Fields.Caller.Key)).Return(json, nil).AnyTimes()
			}
			if config.Encoders.JSON.Fields.Caller.Format != nil {
				json.EXPECT().WithCallerFormat(gomock.Eq(*config.Encoders.JSON.Fields.Caller.Format)).Return(json, nil).AnyTimes()
			}
		}
		if config.Encoders.JSON.Fields.Stacktrace != nil {
			json.EXPECT().WithStacktraceEnabled().Return(json, nil).AnyTimes()
			if config.Encoders.JSON.Fields.Stacktrace.Key != "" {
				json.EXPECT().WithStacktraceKey(gomock.Eq(config.Encoders.JSON.Fields.Stacktrace.Key)).Return(json, nil).AnyTimes()
			}
		}

		framework := &testFramework{
			console: console,
			json:    json,
		}

		writers := make(map[string]*MockWriter)
		if config.Writers.Stdout != nil {
			writers["stdout"] = NewMockWriter(ctrl)
		}
		if config.Writers.Stderr != nil {
			writers["stderr"] = NewMockWriter(ctrl)
		}
		for writerName := range config.Writers.Files {
			writers[writerName] = NewMockWriter(ctrl)
		}

		json.EXPECT().NewWriter(gomock.Any()).DoAndReturn(func(buf io.Writer) (Writer, error) {
			path := buf.(*bytes.Buffer).String()
			writer := writers[path]
			writer.EXPECT().WithSkipCalls(gomock.Eq(2)).Return(writer)
			return writers[path], nil
		}).AnyTimes()
		console.EXPECT().NewWriter(gomock.Any()).DoAndReturn(func(buf io.Writer) (Writer, error) {
			path := buf.(*bytes.Buffer).String()
			writer := writers[path]
			writer.EXPECT().WithSkipCalls(gomock.Eq(2)).Return(writer)
			return writers[path], nil
		}).AnyTimes()

		assert.NoError(t, configure(framework, config, func(path string) (io.Writer, error) {
			return bytes.NewBuffer([]byte(path)), nil
		}))

		for outputName, output := range config.RootLogger.Outputs.Outputs {
			rootLevel := config.RootLogger.Level.Level()
			if output.Level.Level() > rootLevel {
				rootLevel = output.Level.Level()
			}
			writer := writers[outputName]
			if rootLevel.Enabled(DebugLevel) {
				writer.EXPECT().Debug(gomock.Eq("debug"))
			}
			if rootLevel.Enabled(InfoLevel) {
				writer.EXPECT().Info(gomock.Eq("info"))
			}
			if rootLevel.Enabled(WarnLevel) {
				writer.EXPECT().Warn(gomock.Eq("warn"))
			}
			if rootLevel.Enabled(ErrorLevel) {
				writer.EXPECT().Error(gomock.Eq("error"))
			}
			if rootLevel.Enabled(PanicLevel) {
				writer.EXPECT().Panic(gomock.Eq("panic"))
			}
			if rootLevel.Enabled(FatalLevel) {
				writer.EXPECT().Fatal(gomock.Eq("fatal"))
			}
		}

		GetRootLogger().Debug("debug")
		GetRootLogger().Info("info")
		GetRootLogger().Warn("warn")
		GetRootLogger().Error("error")
		GetRootLogger().Panic("panic")
		GetRootLogger().Fatal("fatal")

		loggersNamed := make(map[string]map[string]bool)
		for loggerName, logger := range config.Loggers {
			outputNames := make(map[string]bool)
			for outputName := range config.RootLogger.Outputs.Outputs {
				outputNames[outputName] = true
			}
			for outputName := range logger.Outputs.Outputs {
				outputNames[outputName] = true
			}
			for parentName, parent := range config.Loggers {
				if strings.HasPrefix(loggerName, parentName+"/") {
					for outputName := range parent.Outputs.Outputs {
						outputNames[outputName] = true
					}
				}
			}

			loggerLevel := logger.Level.Level()
			if loggerLevel == EmptyLevel {
				var ancestorNames []string
				ancestorLevels := make(map[string]Level)
				if config.RootLogger.Level.Level() != EmptyLevel {
					ancestorLevels[""] = config.RootLogger.Level.Level()
					ancestorNames = append(ancestorNames, "")
				}
				for ancestorName, ancestor := range config.Loggers {
					if strings.HasPrefix(loggerName, ancestorName+"/") && ancestor.Level.Level() != EmptyLevel {
						ancestorLevels[ancestorName] = ancestor.Level.Level()
						ancestorNames = append(ancestorNames, ancestorName)
					}
				}
				if len(ancestorLevels) > 0 {
					sort.Slice(ancestorNames, func(i, j int) bool {
						return ancestorNames[i] > ancestorNames[j]
					})
					loggerLevel = ancestorLevels[ancestorNames[0]]
				}
			}

			for outputName := range outputNames {
				writer := writers[outputName]

				outputLevel := logger.Outputs.Outputs[outputName].Level.Level()
				if outputLevel == EmptyLevel {
					var ancestorNames []string
					ancestorLevels := make(map[string]Level)
					if output, ok := config.RootLogger.Outputs.Outputs[outputName]; ok && output.Level.Level() != EmptyLevel {
						ancestorLevels[""] = output.Level.Level()
						ancestorNames = append(ancestorNames, "")
					}
					for ancestorName, ancestor := range config.Loggers {
						if strings.HasPrefix(loggerName, ancestorName+"/") {
							if output, ok := ancestor.Outputs.Outputs[outputName]; ok && output.Level.Level() != EmptyLevel {
								ancestorLevels[ancestorName] = output.Level.Level()
								ancestorNames = append(ancestorNames, ancestorName)
							}
						}
					}
					if len(ancestorLevels) > 0 {
						sort.Slice(ancestorNames, func(i, j int) bool {
							return ancestorNames[i] > ancestorNames[j]
						})
						outputLevel = ancestorLevels[ancestorNames[0]]
					}
				}

				if loggerLevel > outputLevel {
					outputLevel = loggerLevel
				}

				outputsNamed, ok := loggersNamed[outputName]
				if !ok {
					outputsNamed = make(map[string]bool)
					loggersNamed[outputName] = outputsNamed
				}
				if !outputsNamed[loggerName] {
					_, ancestorExists := config.RootLogger.Outputs.Outputs[outputName]
					names := strings.Split(loggerName, "/")
					for i := 0; i < len(names); i++ {
						name := strings.Join(names[:i+1], "/")
						if _, ok := config.Loggers[name].Outputs.Outputs[outputName]; ok {
							ancestorExists = true
						}
						if ancestorExists && !outputsNamed[name] {
							writer.EXPECT().WithName(gomock.Eq(name)).Return(writer)
							outputsNamed[name] = true
						}
					}
				}

				if outputLevel.Enabled(DebugLevel) {
					writer.EXPECT().Debug(gomock.Eq("debug"))
				}
				if outputLevel.Enabled(InfoLevel) {
					writer.EXPECT().Info(gomock.Eq("info"))
				}
				if outputLevel.Enabled(WarnLevel) {
					writer.EXPECT().Warn(gomock.Eq("warn"))
				}
				if outputLevel.Enabled(ErrorLevel) {
					writer.EXPECT().Error(gomock.Eq("error"))
				}
				if outputLevel.Enabled(PanicLevel) {
					writer.EXPECT().Panic(gomock.Eq("panic"))
				}
				if outputLevel.Enabled(FatalLevel) {
					writer.EXPECT().Fatal(gomock.Eq("fatal"))
				}
			}
			GetLogger(loggerName).Debug("debug")
			GetLogger(loggerName).Info("info")
			GetLogger(loggerName).Warn("warn")
			GetLogger(loggerName).Error("error")
			GetLogger(loggerName).Panic("panic")
			GetLogger(loggerName).Fatal("fatal")
		}
	})
}

func getLoggingConfig(data []byte) (loggingConfig, error) {
	var config loggingConfig

	consumer := fuzz.NewConsumer(data)
	consoleEncoderEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if consoleEncoderEnabled {
		consoleEncoderFields, err := getEncoderFieldsConfig(consumer, false)
		if err != nil {
			return config, err
		}
		config.Encoders.Console.Fields = consoleEncoderFields
	}

	jsonEncoderEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if jsonEncoderEnabled {
		jsonEncoderFields, err := getEncoderFieldsConfig(consumer, true)
		if err != nil {
			return config, err
		}
		config.Encoders.JSON.Fields = jsonEncoderFields
	}

	stdoutWriterEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if stdoutWriterEnabled {
		config.Writers.Stdout = &stdoutWriterConfig{}
		encodingIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		encoding := encodings[encodingIndex%len(encodings)]
		config.Writers.Stdout.Encoder = encoding
	}

	stderrWriterEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if stderrWriterEnabled {
		config.Writers.Stderr = &stderrWriterConfig{}
		encodingIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		encoding := encodings[encodingIndex%len(encodings)]
		config.Writers.Stderr.Encoder = encoding
	}

	numFileWriters, err := consumer.GetInt()
	if err != nil {
		return config, err
	}
	config.Writers.Files = make(map[string]fileWriterConfig)
	for i := 0; i < numFileWriters%maxNumFileWriters; i++ {
		fileName, err := consumer.GetStringFrom(alphabetChars, 12)
		if err != nil {
			return config, err
		}
		encodingIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		encoding := encodings[encodingIndex%len(encodings)]
		config.Writers.Files[fileName] = fileWriterConfig{
			writerConfig: writerConfig{
				Encoder: encoding,
			},
			Path: fileName,
		}
	}

	rootLoggerConfig, err := getLoggerConfig(consumer, config)
	if err != nil {
		return config, err
	}
	config.RootLogger = rootLoggerConfig

	numLoggers, err := consumer.GetInt()
	if err != nil {
		return config, err
	}
	config.Loggers = make(map[string]loggerConfig)
	for i := 0; i < numLoggers%maxNumLoggers; i++ {
		loggerName, err := consumer.GetStringFrom(loggerNameChars, 64)
		if err != nil {
			return config, err
		}

		loggerConfig, err := getLoggerConfig(consumer, config)
		if err != nil {
			return config, err
		}
		config.Loggers[loggerName] = loggerConfig
	}
	return config, nil
}

func getEncoderFieldsConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (encoderFieldsConfig, error) {
	var config encoderFieldsConfig

	messageField, err := getMessageEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Message = messageField

	nameField, err := getNameEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Name = nameField

	levelField, err := getLevelEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Level = levelField

	timeField, err := getTimestampEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Time = timeField

	callerField, err := getCallerEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Caller = callerField

	stacktraceField, err := getStacktraceEncoderConfig(consumer, includeKeys)
	if err != nil {
		return config, err
	}
	config.Stacktrace = stacktraceField
	return config, nil
}

func getMessageEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*messageEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &messageEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	return config, nil
}

func getNameEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*nameEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &nameEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	return config, nil
}

func getLevelEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*levelEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &levelEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	formatEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if formatEnabled {
		levelFormatIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		levelFormat := levelFormats[levelFormatIndex%len(levelFormats)]
		config.Format = &levelFormat
	}
	return config, nil
}

func getTimestampEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*timestampEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &timestampEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	formatEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if formatEnabled {
		levelFormatIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		timestampFormat := timestampFormats[levelFormatIndex%len(levelFormats)]
		config.Format = &timestampFormat
	}
	return config, nil
}

func getCallerEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*callerEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &callerEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	formatEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if formatEnabled {
		levelFormatIndex, err := consumer.GetInt()
		if err != nil {
			return config, err
		}
		callerFormat := callerFormats[levelFormatIndex%len(levelFormats)]
		config.Format = &callerFormat
	}
	return config, nil
}

func getStacktraceEncoderConfig(consumer *fuzz.ConsumeFuzzer, includeKeys bool) (*stacktraceEncoderConfig, error) {
	fieldEnabled, err := consumer.GetBool()
	if err != nil {
		return nil, err
	}
	if !fieldEnabled {
		return nil, nil
	}
	config := &stacktraceEncoderConfig{}
	if !includeKeys {
		return config, nil
	}
	keyEnabled, err := consumer.GetBool()
	if err != nil {
		return config, err
	}
	if keyEnabled {
		key, err := consumer.GetStringFrom(alphabetChars, 6)
		if err != nil {
			return config, err
		}
		config.Key = key
	}
	return config, nil
}

func getLoggerConfig(consumer *fuzz.ConsumeFuzzer, config loggingConfig) (loggerConfig, error) {
	var logger loggerConfig

	levelEnabled, err := consumer.GetBool()
	if err != nil {
		return logger, err
	}
	if levelEnabled {
		levelIndex, err := consumer.GetInt()
		if err != nil {
			return logger, err
		}
		level := loggerLevels[levelIndex%len(loggerLevels)]
		logger.Level = levelConfig(level)
	}

	logger.Outputs.Outputs = make(map[string]outputSchema)
	if config.Writers.Stdout != nil {
		output, ok, err := getOutputConfig(consumer)
		if err != nil {
			return logger, err
		} else if ok {
			logger.Outputs.Outputs["stdout"] = output
		}
	}

	if config.Writers.Stderr != nil {
		output, ok, err := getOutputConfig(consumer)
		if err != nil {
			return logger, err
		} else if ok {
			logger.Outputs.Outputs["stderr"] = output
		}
	}

	for outputName := range config.Writers.Files {
		output, ok, err := getOutputConfig(consumer)
		if err != nil {
			return logger, err
		} else if ok {
			logger.Outputs.Outputs[outputName] = output
		}
	}
	return logger, nil
}

func getOutputConfig(consumer *fuzz.ConsumeFuzzer) (outputSchema, bool, error) {
	var output outputSchema
	outputEnabled, err := consumer.GetBool()
	if err != nil {
		return output, false, err
	}
	if !outputEnabled {
		return output, false, nil
	}
	outputLevelEnabled, err := consumer.GetBool()
	if err != nil {
		return output, false, err
	}
	if outputLevelEnabled {
		outputLevelIndex, err := consumer.GetInt()
		if err != nil {
			return output, false, err
		}
		outputLevel := loggerLevels[outputLevelIndex%len(loggerLevels)]
		output.Level = levelConfig(outputLevel)
	}
	return output, true, nil
}
