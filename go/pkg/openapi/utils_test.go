// SPDX-FileCopyrightText: (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExtractAllPaths(t *testing.T) {
	spec := &openapi3.T{
		Paths: openapi3.Paths{
			"/put-test-path": &openapi3.PathItem{
				Put: &openapi3.Operation{},
			},
			"/post-test-path": &openapi3.PathItem{
				Post: &openapi3.Operation{},
			},
			"/get-test-path": &openapi3.PathItem{
				Get: &openapi3.Operation{},
			},
			"/delete-test-path": &openapi3.PathItem{
				Delete: &openapi3.Operation{},
			},
		},
	}

	// Test extracting paths that require the "POST" verb
	allPaths := ExtractAllPaths(spec)
	putPaths := allPaths["PUT"]
	expectedPutPaths := []string{"/put-test-path"}
	assert.Equal(t, len(expectedPutPaths), len(putPaths))
	assert.Equal(t, expectedPutPaths, putPaths)

	postPaths := allPaths["POST"]
	expectedPostPaths := []string{"/post-test-path"}
	assert.Equal(t, len(expectedPostPaths), len(postPaths))
	assert.Equal(t, expectedPostPaths, postPaths)

	deletePaths := allPaths["DELETE"]
	expectedDeletePaths := []string{"/delete-test-path"}
	assert.Equal(t, len(expectedDeletePaths), len(deletePaths))
	assert.Equal(t, expectedDeletePaths, deletePaths)

	getPaths := allPaths["GET"]
	expectedGetPaths := []string{"/get-test-path"}
	assert.Equal(t, len(expectedGetPaths), len(getPaths))
	assert.Equal(t, expectedGetPaths, getPaths)

}

func TestLoadOpenAPISpec(t *testing.T) {
	// Create a temporary OpenAPI file
	file, err := os.CreateTemp("", "openapi*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.RemoveAll(file.Name())

	// Write a dummy OpenAPI document to the file
	_, err = file.Write([]byte(`openapi: 3.0.0
info:
  title: Test API`))
	if err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}

	// Call the LoadOpenAPISpec method with the temporary file name
	spec, err := LoadOpenAPISpec(file.Name())
	assert.NoError(t, err)
	assert.Equal(t, "3.0.0", spec.OpenAPI, "unexpected OpenAPI version")
	assert.Equal(t, "Test API", spec.Info.Title, "unexpected API title")
}

func TestLoadOpenAPIYamlUnmarshallingError(t *testing.T) {
	invalidYaml := []byte("invalid: yaml: content")
	tempFile, err := os.CreateTemp("", "invalid_openapi_*.yaml")
	if err != nil {
		t.Fatalf("Could not create temporary file: %v", err)
	}
	defer os.RemoveAll(tempFile.Name())

	_, err = tempFile.Write(invalidYaml)
	assert.NoError(t, err, "error ins writing invalid yaml file into a file")

	// Load the invalid YAML file using LoadOpenAPI
	_, err = LoadOpenAPISpec(tempFile.Name())
	assert.Error(t, err, "LoadOpenAPI should return an error for invalid YAML content")
}

// TestLoadOpenAPINonExistentFile tests the LoadOpenAPI function for a non-existent file
func TestLoadOpenAPINonExistentFile(t *testing.T) {
	nonExistentFile := "non_existent_file.yaml"

	// Load the non-existent file using LoadOpenAPI
	_, err := LoadOpenAPISpec(nonExistentFile)
	assert.Error(t, err, "LoadOpenAPI should return an error for non-existent file")
}
