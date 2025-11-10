// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/orch-library/go/pkg/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// CatalogV3Upload represents the upload artifact structure
type CatalogV3Upload struct {
	Artifact []byte `json:"artifact"`
	FileName string `json:"fileName"`
}

// CatalogV3UploadRequest represents the full request structure for catalog upload API
type CatalogV3UploadRequest struct {
	Upload       *CatalogV3Upload `json:"upload,omitempty"`
	SessionID    *string          `json:"sessionId,omitempty"`
	LastUpload   *bool            `json:"lastUpload,omitempty"`
	UploadNumber *int             `json:"uploadNumber,omitempty"`
}

// CatalogV3UploadResponse represents the response from catalog upload API
type CatalogV3UploadResponse struct {
	ErrorMessages *[]string `json:"errorMessages,omitempty"`
	SessionID     string    `json:"sessionId"`
	UploadNumber  int       `json:"uploadNumber"`
}

// LoadResources loads the specified files or directories as catalog resources using the V3 API
func (l *Loader) LoadResources(ctx context.Context, accessToken string, paths []string) error {
	url := fmt.Sprintf("%s/v3/projects/%s/catalog/uploads", l.catalogEndpoint, l.projectName)

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

	// Upload each file individually using CatalogV3Upload format
	client := &http.Client{}
	var sessionID *string
	var allErrors []string
	totalFiles := len(filesToUpload)

	for i, file := range filesToUpload {
		// Create CatalogV3Upload request body
		upload := CatalogV3Upload{
			Artifact: file.content,
			FileName: file.fileName,
		}

		// Determine if this is the last upload
		isLastUpload := (i == totalFiles-1)
		lastUpload := isLastUpload

		// Build the full request structure
		uploadRequest := CatalogV3UploadRequest{
			Upload:     &upload,
			LastUpload: &lastUpload,
		}

		// Add sessionId for all uploads after the first one
		if sessionID != nil {
			uploadRequest.SessionID = sessionID
		}

		// Add upload number (1-indexed)
		uploadNum := i + 1
		uploadRequest.UploadNumber = &uploadNum

		jsonData, err := json.Marshal(uploadRequest)
		if err != nil {
			return err
		}

		r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err := client.Do(r)
		if err != nil {
			return err
		}

		// Parse response
		var uploadResp CatalogV3UploadResponse
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return errors.NewInvalid("upload failed for %s: %d: %s", file.fileName, resp.StatusCode, string(body))
		}

		if err := json.Unmarshal(body, &uploadResp); err != nil {
			return errors.NewInvalid("failed to parse upload response for %s: %v", file.fileName, err)
		}

		// Store session ID from first upload for subsequent requests
		if sessionID == nil && uploadResp.SessionID != "" {
			sessionID = &uploadResp.SessionID
		}

		// Collect any error messages
		if uploadResp.ErrorMessages != nil && len(*uploadResp.ErrorMessages) > 0 {
			allErrors = append(allErrors, *uploadResp.ErrorMessages...)
		}
	}

	// Return error if there were any error messages
	if len(allErrors) > 0 {
		return errors.NewInvalid("upload completed with errors: %s", strings.Join(allErrors, "\n"))
	}

	return nil
}

// IsDir Returns true if the given path is a directory
func IsDir(path string) (bool, error) {
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return false, err
	}
	stat, err := file.Stat()
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}
