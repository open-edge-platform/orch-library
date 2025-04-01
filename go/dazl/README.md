<!---
  SPDX-FileCopyrightText: (C) 2024 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# dazl

## Configurable abstraction layer for Go logging frameworks

Dazl is not just another Go logging framework. We're not here to reinvent Go logging for the nth time. Dazl is an
abstraction layer that provides a unified interface and configuration format on top of existing logging frameworks,
using a proven approach adapted from popular frameworks in other languages like [slf4j](https://slf4j.org).

Dazl is a pluggable logging abstraction with support for multiple existing backend frameworks:

- [zap](https://github.com/uber-go/zap)
- [zerolog](https://github.com/rs/zerolog)

Dazl loggers add numerous features on top of existing frameworks:

- Decouples Go libraries from specific logging implementations
- Makes logging configurable via YAML configuration files
- Structured logging with support for JSON or console encoding and user-defined fields
- Hierarchical loggers, inheritance, sampling and other advanced features
- Supports runtime configuration changes for easy debugging

# User Guide

- [Use cases](#use-cases)
  - [Logging in Go libraries](#logging-in-go-libraries)
  - [Configuration of Go services](#configuration-of-go-application-logging)
- [Getting started](#getting-started)
  - [Add the dazl dependency](#add-the-dazl-dependencies)
  - [Import a logging backend](#import-a-logging-backend)
  - [Configure logging](#configure-logging)
  - [Run the application](#run-the-application)
- [Go API](#go-api)
  - [Initializing the framework](#initializing-the-logging-framework)
  - [Working with loggers](#loggers)
  - [Log levels](#log-levels)
  - [Structured logging](#structured-logging)
- [Configuration files](#configuration-files)
  - [Encoders](#configuring-encoders)
    - [JSON](#json-encoder)
    - [Console](#console-encoder)
  - [Writers](#configuring-writers)
  - [Loggers](#configuring-loggers)
  - [Reference](#reference)
- [Runtime configuration changes](#runtime-configuration-changes)
  - [Changing the log level](#changing-the-log-level)
- [Custom logging frameworks](#custom-logging-frameworks)
  - [Custom encoders](#encoding)
  - [Custom writers](#log-writers)

## Use cases

There are numerous benefits to using a logging abstraction, but dazl is designed to serve a couple of important use
cases for two types of applications in particular.

### Logging in Go libraries

Go libraries designed to be imported and used by other Go modules can use dazl to avoid adding a dependency on a
specific logging framework, tying their users to the same framework. Additionally, dazl enables your users to configure
loggers and log formats indepenedently, with no added work for you. The users of your library ought to be able to
select their own logging framework and configure the format and severity of log outputs. Simply add a dependency on
the dazl logger, and leave it up to your users to import whichever logging backend they desire.

### Configuration of Go application logging

One of the most common use cases for Go applications is in cloud applications deployed in containers and on platforms
like Kubernetes. Most Go logging frameworks provide programmatic APIs for configuring loggers, levels, formats, and
other logging options. Using dazl in Go-based services enables your users to configure logging independent of the
code, eliminating the need to recompile code to modify the verbosity or format of application logs.

# Getting started

## Add the dazl dependencies

To start using dazl, first add the framework to your `go.mod`:

```bash
go get -u github.com/atomix/dazl
```

Additionally, add the dependency for one of the [logging backends](#initializing-the-logging-framework)L

```bash
go get -u github.com/atomix/dazl/zap
```

## Import a logging backend

Import the logging backend in your module `main`:

`main.go`

```go
package main

import (
    "github.com/atomix/dazl"
    _ "github.com/atomix/dazl/zap"
)

var log = dazl.GetLogger()

func main() {
    log.Info("Hello world!")
}
```

To avoid adding unnecessary dependencies on logging frameworks, logging backends should only be imported within
your module `main`.

## Configure logging

Create a `logging.yaml` file to configure the logging framework:

`logging.yaml`

```yaml
encoders:
  json:
    fields:
      - message
      - level:
          format: uppercase
      - caller:
          format: short
      - timestamp:
          format: iso8601

writers:
  stdout:
    encoder: json

rootLogger:
  level: info
  outputs:
    - stdout
```

To print logs to an output, the `logging.yaml` configuration must configure one or more `writers` and configure the
`rootLogger` or child `loggers` with `outputs` to output messages from those loggers to the configured writer(s).

## Run the application

Finally, when you run the application `main` from the same directory as your `logging.yaml` file, you should see
the logs printed to stdout:

```bash
go run ./main.go
```

For more information [check out the example module](./examples).

# Go API

## Initializing the logging framework

The dazl logger does not log to any output unless a logging framework is imported. To maintain independence from any
particular logging backend, applications should only import a specific logging framework from within a `main` file.
Libraries designed to be imported by other projects should never import a logging backend themselves. Instead, leave
the specific logging framework implementation up to your users.

The logging backend is configured by importing the framework into your application's `main` package.

### Logging with zap

To configure dazl to use the [zap](https://github.com/uber-go/zap) logging backend, add the `zap` framework
to your module's `go.mod`:

```bash
go get -u github.com/atomix/dazl/zap
```

Then import the `github.com/atomix/dazl/zap` framework implementation in your `main` package:

```go
package main

import _ "github.com/atomix/dazl/zap"

func main() {
    ...
}
```

### Logging with zerolog

To configure dazl to use the [zerolog](https://github.com/rs/zerolog) logging backend, add the `zerolog` framework
to your module's `go.mod`:

```bash
go get -u github.com/atomix/dazl/zerolog
```

Then import the `github.com/atomix/dazl/zerolog` framework implementation in your `main` package:

```go
package main

import _ "github.com/atomix/dazl/zerolog"

func main() {
    ...
}
```

## Loggers

The typical usage of the framework is to create a `Logger` once at the top of each Go package:

```go
var log = dazl.GetLogger()
```

Package loggers are assigned the name of the package calling `GetLogger()` or `GetPackageLogger()`. So, if you
call `dazl.GetLogger()` from the `github.com/atomix/atomix/runtime` package, the logger will be assigned the
name `github.com/atomix/atomix/runtime`. The naming strategy becomes important for
[logger configuration](#configuration-files) and, in particular, [inheritance](#inheritance).

```go
const author = "kuujo"

var log = dazl.GetLogger()

func main() {
    log.Infof("The author of dazl is %s", author)
}
```

```text
2023-03-31T01:13:30.607Z	INFO	main.go:12	The author of dazl is kuujo
```

Alternatively, custom loggers can be retrieved via `GetLogger`:

```go
var log = dazl.GetLogger("foo/bar")
```

Logger names must be formatted in path format, with each element separated by a `/`. This format is used
to establish a hierarchy for [inheritence](#inheritance) of logger configurations.

All loggers are descendants of the root logger:

```go
var log = dazl.GetRootLogger()
```

## Log levels

Dazl supports a fairly standard set of log levels for loggers:

- `debug`
- `info`
- `warn`
- `error`
- `fatal`
- `panic`

The levels for each logger can be configured individually via their configuration:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    level: warn
```

The `Logger` interface exposes methods for simple logging, formatted logging, and structured logging with typed
fields for each log level:

- `Debug(args ...any)`
- `Debugf(msg string, args ...any)`
- `Debugw(msg string, fields ...Field)`
- `Info(args ...any)`
- `Infof(msg string, args ...any)`
- `Infow(msg string, fields ...Field)`
- `Warn(args ...any)`
- `Warnf(msg string, args ...any)`
- `Warnw(msg string, fields ...Field)`
- ...

Messages will only be written to log [outputs](#outputs) if the configured level of the logger is higher than the
message level.

## Structured logging

Structured logging is supported for the JSON [encoding](#configuring-encoders), and JSON fields are configurable via
the `Logger` API.

The simplest way to add fields to your structured logs is to call one of the `-w` methods on the `Logger` interface.
These methods accept an arbitrary number of `...Field`s to write to the logs. Fields are typed and named values
that can be constructed via functions in the `dazl` package:

```go
log.Warnw("Something went wrong!", 
    dazl.String("user", user.Name), 
    dazl.Uint64("user-id", user.ID))
```

Alternatively, you can create a structured logger with a fixed set of fields using the `WithFields` method:

```go
var log = dazl.GetLogger().WithFields(
    dazl.String("user", user.Name),
    dazl.Uint64("id", user.ID))
log.Warn("Something went wrong!")
```

When the logger is output to a JSON encoded writer, the above code will log the fields as part of the JSON object:

```text
{"timestamp":"2023-04-07T19:24:09-07:00","logger":"2/4","message":"Something went wrong!","user":"Jordan Halterman","id":5678}
```

# Configuration files

Loggers can be configured via a YAML configuration file. The configuration files may be in one of many
locations on the file system:

- `logging.yaml`
- `~/logging.yaml`
- `/etc/dazl/logging.yaml`

The configuration file contains a set of `loggers` which specifies the level and outputs of each logger,
`writers` which specify where to write log messages, and `encoders` defining how to encode log messages.

## Configuring encoders

The `encoders` section of the configuration defines how dazl encodes log messages. Dazl supports two
encodings: `json` and `console`. Each encoder defines the set of `fields` to output and optionally the
format of the fields.

```yaml
encoders:
  json:
    fields:
      ...
  console:
    fields:
      ...
```

Encoders are referenced by [`writers`](#configuring-writers) and used to encode messages.

Note that support for some encoding options such as renaming keys or formatting fields depends on whether those
features are supported the underlying logging framework. If some requested features are not supported by the
imported logging framework, dazl may panic at startup.

### JSON encoder

The `json` encoder configuration defines the set of fields to include in all JSON formatted messages. Each
JSON field also supports an additional `key` to override the default JSON key for that field:

```yaml
encoders:
  json:
    # A list of fields to include
    fields:
      # The log message
      - message:
          # The JSON key for the field
          key: msg
      # The logger name
      - name:
          # The JSON key for the field
          key: logger
      # The log level
      - level:
          # The JSON key for the field
          key: level
          # The level format: 'uppercase' or 'lowercase'
          format: uppercase
      # The time at which the message was logged
      - timestamp:
          # The JSON key for the field
          key: time
          # The time format: 'iso8601' or 'unix'
          format: iso8601
      # The line of code at which the message was logged
      - caller:
          # The JSON key for the field
          key: caller
          # The caller format: 'short' or 'long'
          format: short
      # The log stacktrace
      - stacktrace:
          # The JSON key for the field
          key: trace
```

### Console encoder

The `console` encoder configuration defines the colums to encode with each log message:

```yaml
encoders:
  console:
    # A list of fields to include
    fields:
      # The log message
      - message
      # The logger name
      - name
      # The log level
      - level:
          # The level format: 'uppercase' or 'lowercase'
          format: uppercase
      # The time at which the message was logged
      - timestamp:
          # The time format: 'iso8601' or 'unix'
          format: iso8601
      # The line of code at which the message was logged
      - caller:
          # The caller format: 'short' or 'long'
          format: short
      # The log stacktrace
      - stacktrace
```

## Encoder fields

Both the `json` and `console` encoders support the following set of fields:

- `message`
- `name`
- `level`
- `timestamp`
- `caller`
- `stacktrace`

If any field is excluded from the encoder's field list, dazl will _attempt_ to exclude that field from the log
output, but some logging backends may not support this ability and therefore may include those fields in their output.

### Level formats

The format of the log level can be configured for each encoder via the `format` key:

```yaml
encoders:
  json:
    level:
      format: uppercase
```

Defined level formats include:

- `uppercase` - upper case level name
- `lowercase` - lower case level name

Note that support for level formats depends on support from the imported logging backend. Dazl may panic at startup
if the underlying logging framework does not support the configured level format.

### Timestamp formats

The format of the timestamp can be configured for each encoder via the `format` key:

```yaml
encoders:
  json:
    timestamp:
      format: ISO8601
```

Defined timestamp formats include:

- `ISO8601`
- `unix`

Note that support for timestamp formats depends on support from the imported logging backend. Dazl may panic at startup
if the underlying logging framework does not support the configured timestamp format.

### Caller formats

The format of the caller can be configured for each encoder via the `format` key:

```yaml
encoders:
  json:
    caller:
      format: uppercase
```

Defined caller formats include:

- `short`
- `full`

Note that support for caller formats depends on support from the imported logging backend. Dazl may panic at startup
if the underlying logging framework does not support the configured level format.

## Configuring writers

```yaml
# A set of named writers for loggers to write to.
# All writers must specify an 'encoder' to use
writers:
  # The stdout writer
  stdout:
    # The name of the encoder to use for the writer
    encoder: console
  # The stderr writer
  stderr:
    # The name of the encoder to use for the writer
    encoder: console
  # Remaining writers are files
  file:
    # The path to the file
    path: ./example.log
    # The name of the encoder to use for the writer
    encoder: json
```

## Configuring loggers

### The root logger

The default logging configuration is configured via the `rootLogger` key:

```yaml
rootLogger:
  level: info
```

All loggers inherit their default configuration from the root logger. The root logger configuration should at
least specify a minimum log `level` for all loggers, and at least one `outputs` to a [writer](#configuring-writers).

```yaml
rootLogger:
  level: info
  outputs:
    - stdout
```

### Named loggers

Once the `rootLogger` has been defined, all other loggers that are descendants of the root logger may be defined and
configured in the `loggers` section of the configuration file:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    level: info
  github.com/atomix/atomix/sidecar:
    level: warn
```

### Outputs

Once you've defined the set of writers to which to write your application logs, named loggers can be directed to
those writers via their configured `outputs`:

```yaml
rootLogger:
  level: info
  outputs:
    - stdout
```

Each output must specify a `writer` to write to. As with log levels, loggers inherit the outputs of their ancestors, so
with an output to the `stdout` writer in the root logger configuration, all loggers will have an output to `stdout`.

### Output levels

In some cases, you may want to restrict the verbosity of logs to one output without restricting the verbosity of all
messages for a logger. For example, you may want to write `info` and higher messages to the console for
human readability, and `debug` and higher messages to a file for later debugging. Each output supports a `level`
that can be used to filter the logs to that output:

```yaml
loggers:
  github.com/atomix/atomix:
    # Pass through all messages over 'debug' level to the outputs
    level: debug
    outputs:
      # Limit the messages from this logger to stdout to 'info' level
      - stdout:
          level: info
      # Pass through all messages to the 'file' writer
      - file

writers:
  stdout:
    encoder: console
  file:
    path: ./app.log
    encoder: json
```

### Output overrides

Descendants may override their ancestor loggers' output configurations. This can be done by simply specifying the
same output name. For example, to override the log level for the `rootLogger`'s `stdout` output:

```yaml
# Configure the stdout writer to use console encoding
writers:
  stdout:
    encoder: console

rootLogger:
  level: info
  outputs:
    # Output the root logger to stdout
    - stdout

loggers:
  github.com/atomix/atomix/runtime:
    outputs:
      # Limit outputs from this logger to stdout to minimum of 'warn' level
      stdout:
        level: warn
```

### Sampling

Samplers can be added to either loggers to reduce the number of messages logged:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    sample:
      basic:
        interval: 10
```

Samplers can also be added to individual outputs to limit only the messages to that particular output:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    outputs:
      stdout:
        sample:
          basic:
            interval: 10
```

### Basic sampler

The `basic` sampler logs every nth message below `maxLevel` according to the configured `interval`:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    sample:
      basic:
        interval: 10
        maxLevel: debug
```

### Random sampler

The `random` sampler randomly logs messages below `maxLevel` by choosing a random integer between `0` and `interval`:

```yaml
loggers:
  github.com/atomix/atomix/runtime:
    sample:
      random:
        interval: 10
        maxLevel: debug
```

### Inheritance

The path-like format used for logger names is used to establish a hierarchy of loggers. The dazl configuration
enables developers and their users to configure individual loggers at runtime. Log levels are inherited by
descendants of a logger. This enables users to easily enable logging for specific Go packages, their subpackages,
or entire Go modules with a single configuration change:

```yaml
# Enable debug logging for all Atomix code
loggers:
  github.com/atomix:
    level: debug
```

You can set the `dazl.Level` for a logger at startup time via configuration files or at runtime
via the `Logger` API to control the granularity of a logger's output:

```go
dazl.GetLogger("github.com/atomix").SetLevel(dazl.DebugLevel)
```

The root logger is the ancestor of all other loggers and can be configured via `GetRootLogger`:

```go
dazl.GetRootLogger().SetLevel(dazl.InfoLevel)
```

If the level for a logger is not explicitly set, it will inherit its level from its nearest ancestor in
the logger hierarchy. For example, setting the `github.com/atomix/atomix/runtime` logger to the `debug`
level will change the loggers for all loggers in the `github.com/atomix/atomix/runtime/...` packages
to the `debug` level.

## Reference

A [reference configuration file](./examples/reference.yaml) detailing and documenting all the available configuration
options in `logging.yaml` is available in the [examples](./examples) directory of this repo.

# Runtime configuration changes

## Changing the log level

You can set the `dazl.Level` for a logger at startup time via configuration files or at runtime
via the `Logger` API to control the granularity of a logger's output:

```go
dazl.GetLogger("github.com/atomix").SetLevel(dazl.DebugLevel)
```

The root logger is the ancestor of all other loggers and can be configured via `GetRootLogger`:

```go
dazl.GetRootLogger().SetLevel(dazl.InfoLevel)
```

# Custom logging frameworks

Dazl provides several existing implementations of logging frameworks:

- [zap](./zap/framework.go)
- [zerolog](./zerolog/framework.go)

Logging frameworks are implemented by implementing the `Framework` interface:

```go
const name = "example"

type Framework struct{}

func (f Framework) Name() string {
    return name
}
```

Frameworks should register when imported:

```go
func init() {
    dazl.Register(&Framework{})
}
```

## Encoding

`Framework` implementations should implement one or more of the `*EncodingFramework` interfaces to indicate support
for encoding formats. Frameworks implement these interfaces to provide `Encoder`s which are used by dazl to create new
`Writer`s:

```go
type JSONEncoder struct{}

func (e *JSONEncoder) NewWriter(writer io.Writer) (dazl.Writer, error) {
    ...
}
```

## JSON encoding

To implement support for JSON encoding, implement the `JSONEncodingFramework` interface:

```go
func (f Framework) JSONEncoder() Encoder {
    return &JSONEncoder{}
}
```

## Console encoding

To implement support for JSON encoding, implement the `ConsoleEncodingFramework` interface:

```go
func (f Framework) ConsoleEncoder() Encoder {
    return &ConsoleEncoder{}
}
```

## Encoding options

`Encoder` implementations may support configuration options by implementing optional interfaces with the
following methods:

- `WithMessageKey(key string) (dazl.Encoder, error)`
- `WithNameEnabled() (dazl.Encoder, error)`
- `WithNameKey(key string) (dazl.Encoder, error)`
- `WithLevelEnabled() (dazl.Encoder, error)`
- `WithLevelKey(key string) (dazl.Encoder, error)`
- `WithLevelFormat(format dazl.LevelFormat) (dazl.Encoder, error)`
- `WithTimestampEnabled() (dazl.Encoder, error)`
- `WithTimestampKey(key string) (dazl.Encoder, error)`
- `WithTimestampFormat(format dazl.TimestampFormat) (dazl.Encoder, error)`
- `WithCallerEnabled() (dazl.Encoder, error)`
- `WithCallerKey(key string) (dazl.Encoder, error)`
- `WithCallerFormat(format dazl.CallerFormat) (dazl.Encoder, error)`
- `WithStacktraceEnabled() (dazl.Encoder, error)`
- `WithStacktraceKey(key string) (dazl.Encoder, error)`

## Log writers

`Encoder`s create `Writer`s for use by the dazl `Logger`. All `Writer`s must implement the following methods:

- `WithName(name string) dazl.Writer`
- `WithSkipCalls(calls int) dazl.Writer`
- `Debug(msg string)`
- `Info(msg string)`
- `Warn(msg string)`
- `Error(msg string)`
- `Panic(msg string)`
- `Fatal(msg string)`

### Writer options

`Writer` implementations may optionally support additional features by implementing optional interfaces
by adding the following methods:

- `WithStringField(name string, value string) dazl.Writer`
- `WithIntField(name string, value int) dazl.Writer`
- `WithInt32Field(name string, value int32) dazl.Writer`
- `WithInt64Field(name string, value int64) dazl.Writer`
- `WithUintField(name string, value uint) dazl.Writer`
- `WithUint32Field(name string, value uint32) dazl.Writer`
- `WithUint64Field(name string, value uint64) dazl.Writer`
- `WithFloat32Field(name string, value float32) dazl.Writer`
- `WithFloat64Field(name string, value float64) dazl.Writer`
- `WithBoolField(name string, value bool) dazl.Writer`
- `WithStringSliceField(name string, values []string) dazl.Writer`
- `WithIntSliceField(name string, values []int) dazl.Writer`
- `WithInt32SliceField(name string, values []int32) dazl.Writer`
- `WithInt64SliceField(name string, values []int64) dazl.Writer`
- `WithUintSliceField(name string, values []uint) dazl.Writer`
- `WithUint32SliceField(name string, values []uint32) dazl.Writer`
- `WithUint64SliceField(name string, values []uint64) dazl.Writer`
- `WithFloat32SliceField(name string, values []float32) dazl.Writer`
- `WithFloat64SliceField(name string, values []float64) dazl.Writer`
- `WithBoolSliceField(name string, values []bool) dazl.Writer`
- `WithErrorField(name string, err error) dazl.Writer`
