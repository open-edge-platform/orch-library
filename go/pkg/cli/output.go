// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"
)

const (
	// ExitSuccess means nominal status
	ExitSuccess = iota

	// ExitError means general error
	ExitError

	// ExitBadConnection means failed connection to remote service
	ExitBadConnection

	// ExitBadArgs means invalid argument values were given
	ExitBadArgs = 128
)

var outputWriter io.Writer

// GetOutput returns the current output writer
func GetOutput() io.Writer {
	return outputWriter
}

// CaptureOutput allows a test harness to redirect output to an alternate source for testing
func CaptureOutput(capture io.Writer) {
	outputWriter = capture
}

func init() {
	CaptureOutput(os.Stdout)
}

// Output prints the specified format message with arguments to stdout.
func Output(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(outputWriter, msg, args...)
}
