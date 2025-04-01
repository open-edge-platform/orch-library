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
	"mime/multipart"
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

// LoadResources loads the specified files or directories as catalog resources using the upload API
func (l *Loader) LoadResources(ctx context.Context, accessToken string, paths []string) error {
	url := fmt.Sprintf("%s/v3/projects/%s/catalog/upload", l.catalogEndpoint, l.projectName)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, path := range paths {
		isDirectory, err := IsDir(path)
		if err != nil {
			return err
		}

		if isDirectory {
			dirPath := path
			err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, _ error) error {
				if !d.IsDir() && strings.HasSuffix(path, ".yaml") {
					return addFile(writer, path, dirPath)
				}
				return nil
			})
		} else {
			err = addFile(writer, path, "")
		}
		if err != nil {
			return err
		}
	}

	_ = writer.Close()

	r, _ := http.NewRequestWithContext(ctx, "POST", url, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	if resp, err := client.Do(r); err != nil {
		return err
	} else if resp.StatusCode != 200 {
		messages := extractMessages(resp)
		return errors.NewInvalid("upload failed: %d: %s", resp.StatusCode, messages)
	}
	return nil
}

func addFile(writer *multipart.Writer, path string, dirPath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	part, _ := writer.CreateFormFile("files", strings.TrimPrefix(strings.TrimPrefix(path, dirPath), "/"))
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	return file.Close()
}

func extractMessages(resp *http.Response) string {
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, resp.Body)
	jsonString := buf.String()
	type ResponsesJSON struct {
		Responses []struct {
			SessionID     string   `json:"sessionId"`
			UploadNumber  int      `json:"uploadNumber"`
			ErrorMessages []string `json:"errorMessages"`
		} `json:"responses"`
	}
	var responses ResponsesJSON
	_ = json.Unmarshal([]byte(jsonString), &responses)

	var result = ""
	for _, response := range responses.Responses {
		errorMessages := response.ErrorMessages
		if len(errorMessages) != 0 {
			for _, message := range errorMessages {
				result = result + "\n" + message
			}
		}
	}

	return result
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
