// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-edge-platform/orch-library/go/pkg/errors"
)

// Loader implements the functions used to load YAML files
type Loader struct {
	catalogEndpoint string
	projectName     string
}

// NewLoader returns an initialized loader
func NewLoader(catalogEndpoint string, projectName string) *Loader {
	return &Loader{catalogEndpoint: strings.TrimSuffix(catalogEndpoint, "/"), projectName: projectName}
}

// UploadResponse represents the response from the upload API
type UploadResponse struct {
	Responses []struct {
		SessionID     string   `json:"sessionId"`
		UploadNumber  int      `json:"uploadNumber"`
		ErrorMessages []string `json:"errorMessages"`
	} `json:"responses"`
}

// LoadResources loads the specified files or directories as catalog resources using multipart/form-data
func (l *Loader) LoadResources(ctx context.Context, accessToken string, paths []string) error {
	// Use the correct endpoint: /v3/projects/{project}/catalog/upload (singular, not plural!)
	// This matches what the UI uses and expects multipart/form-data format
	url := fmt.Sprintf("%s/v3/projects/%s/catalog/upload", l.catalogEndpoint, l.projectName)

	// Collect all YAML files to upload
	type fileToUpload struct {
		content  []byte
		fileName string
	}
	var filesToUpload []fileToUpload

	for _, path := range paths {
		isDirectory, err := IsDir(path)
		if err != nil {
			return err
		}

		if isDirectory {
			dirPath := path
			err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, _ error) error {
				if !d.IsDir() && strings.HasSuffix(path, ".yaml") {
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					fileName := strings.TrimPrefix(strings.TrimPrefix(path, dirPath), "/")
					filesToUpload = append(filesToUpload, fileToUpload{
						content:  content,
						fileName: fileName,
					})
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fileName := filepath.Base(path)
			filesToUpload = append(filesToUpload, fileToUpload{
				content:  content,
				fileName: fileName,
			})
		}
	}

	if len(filesToUpload) == 0 {
		return errors.NewInvalid("no YAML files found to upload")
	}

	// Create multipart form with all files
	// The backend expects field name "files" (not "file" or "upload")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range filesToUpload {
		part, err := writer.CreateFormFile("files", file.fileName)
		if err != nil {
			return errors.NewInvalid("failed to create form file for %s: %v", file.fileName, err)
		}
		if _, err := part.Write(file.content); err != nil {
			return errors.NewInvalid("failed to write file content for %s: %v", file.fileName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return errors.NewInvalid("failed to close multipart writer: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return errors.NewInvalid("failed to create HTTP request: %v", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.NewInvalid("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.NewInvalid("failed to read response body: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return errors.NewInvalid("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var uploadResp UploadResponse
	if err := json.Unmarshal(bodyBytes, &uploadResp); err != nil {
		// If we can't parse the response but status was 200, consider it success
		return nil
	}

	// Check for errors in the response
	var allErrors []string
	for _, r := range uploadResp.Responses {
		if len(r.ErrorMessages) > 0 {
			allErrors = append(allErrors, r.ErrorMessages...)
		}
	}

	if len(allErrors) > 0 {
		return errors.NewInvalid("upload completed with errors:\n%s", strings.Join(allErrors, "\n"))
	}

	return nil
}

// IsDir Returns true if the given path is a directory
func IsDir(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
