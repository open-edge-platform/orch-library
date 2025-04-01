// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

// These are mainly placeholders for the moment for functions that are not
// otherwise called
func Test_InitConfig(t *testing.T) {
	InitConfig("test-module")

	assert.Equal(t, "test-module", configName)
}

func Test_AddConfigFlags(_ *testing.T) {
	AddConfigFlags(&cobra.Command{
		Short: "test command",
	}, "localhost:5150")
}

func Test_GetConfigCommand(_ *testing.T) {
	GetConfigCommand()
}

func Test_RunConfigSetCommand(t *testing.T) {
	err := runConfigSetCommand(&cobra.Command{
		Short: "test command",
	}, []string{"a", "b", "c"})
	assert.ErrorContains(t, err, "Config File \"test-module\" Not Found")
}

func Test_GetCertPath(t *testing.T) {
	path := getCertPath(&cobra.Command{
		Short: "test command",
	})
	assert.Len(t, path, 0)
}

func Test_Config(t *testing.T) {
	var cmd *cobra.Command
	var err error
	dir, err := os.MkdirTemp("", "*")
	assert.NoError(t, err)
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)

	var output strings.Builder
	CaptureOutput(&output)

	getArgs := []string{"abc"}
	setArgs := []string{"abc", "123"}

	// test profile creation
	homedir.DisableCache = true
	SetConfigDir("CLI-TEST")
	InitConfig("cli-test")

	cmd = getConfigInitCommand()
	err = cmd.RunE(cmd, nil)
	assert.NoError(t, err)

	// Make sure the default config was created
	_, err = os.Stat(dir + "/CLI-TEST/cli-test.yaml")
	assert.NoError(t, err)

	// test empty get
	output.Reset()
	cmd = getConfigGetCommand()
	assert.NotNil(t, cmd)
	err = cmd.RunE(cmd, getArgs)
	assert.NoError(t, err)
	assert.Equal(t, "<nil>\n", output.String())

	// test set
	output.Reset()
	cmd = getConfigSetCommand()
	assert.NotNil(t, cmd)
	err = cmd.RunE(cmd, setArgs)
	assert.NoError(t, err)
	assert.Equal(t, "123\n", output.String())

	// test get with value set
	output.Reset()
	cmd = getConfigGetCommand()
	assert.NotNil(t, cmd)
	err = cmd.RunE(cmd, getArgs)
	assert.NoError(t, err)
	assert.Equal(t, "123\n", output.String())

	// test delete
	output.Reset()
	cmd = getConfigDeleteCommand()
	assert.NotNil(t, cmd)
	err = cmd.RunE(cmd, getArgs)
	assert.NoError(t, err)
	assert.Equal(t, "<nil>\n", output.String())
}

func Test_Output(t *testing.T) {
	var output strings.Builder
	CaptureOutput(&output)
	Output("%s %s", "hello", "there")
	assert.Equal(t, "hello there", output.String())
}
