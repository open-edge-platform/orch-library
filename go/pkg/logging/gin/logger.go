// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/gin-gonic/gin"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"io"
	"strings"
	"time"
)

// WriteFunc convert func to io.Writer.
type WriteFunc func([]byte) (int, error)

// Write overwrites write method for gin logger
func (fn WriteFunc) Write(data []byte) (int, error) {
	return fn(data)
}

// GinLogger a custom gin logger using dazl logger
type GinLogger struct {
	log dazl.Logger
}

func NewWriter(log dazl.Logger) io.Writer {
	return WriteFunc(func(data []byte) (int, error) {
		msg := strings.TrimSpace(string(data))
		if strings.Contains(msg, "[GIN-debug]") {
			log.Debug(msg)
		} else if strings.Contains(msg, "[WARNING]") {
			log.Warn(msg)
		} else if strings.Contains(msg, "[ERROR]") {
			log.Error(msg)
		} else {
			log.Info(msg)
		}
		return 0, nil
	})
}

// NewGinLogger returns a gin.HandlerFunc (middleware) that logs requests using dazl
//
// Requests with errors are logged using dazl.Errorw().
// Requests without errors are logged using dazl.Infow().
func NewGinLogger(logger dazl.Logger) gin.HandlerFunc {
	ginLogger := GinLogger{
		log: logger,
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		fields := []dazl.Field{
			dazl.Int("status", c.Writer.Status()),
			dazl.String("method", c.Request.Method),
			dazl.String("path", path),
			dazl.String("query", query),
			dazl.String("ip", c.ClientIP()),
			dazl.String("user-agent", c.Request.UserAgent()),
			dazl.Duration("latency (ns)", latency),
		}

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				ginLogger.log.Errorw(e, fields...)
			}
		} else {
			ginLogger.log.Infow(path, fields...)
		}

	}
}
