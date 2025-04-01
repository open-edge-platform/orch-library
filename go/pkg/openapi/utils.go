// SPDX-FileCopyrightText: (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"

	"gopkg.in/yaml.v3"
	"os"
)

const (
	PUT    = "PUT"
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

func LoadOpenAPISpec(filename string) (*openapi3.T, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec *openapi3.T
	err = yaml.Unmarshal(bytes, &spec)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func ExtractAllPaths(spec *openapi3.T) map[string][]string {
	paths := make(map[string][]string)
	for path, item := range spec.Paths {
		if item.Post != nil {
			paths[POST] = append(paths[POST], path)
		}
		if item.Delete != nil {
			paths[DELETE] = append(paths[DELETE], path)

		}
		if item.Get != nil {
			paths[GET] = append(paths[GET], path)

		}
		if item.Put != nil {
			paths[PUT] = append(paths[PUT], path)
		}
	}
	return paths
}
