// SPDX-FileCopyrightText: (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRoutingError(t *testing.T) {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	marshaler := &runtime.JSONPb{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	tests := []struct {
		name       string
		httpStatus int
		wantStatus int
	}{
		{
			name:       "StatusMethodNotAllowed",
			httpStatus: http.StatusMethodNotAllowed,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "StatusNotFound",
			httpStatus: http.StatusNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w = httptest.NewRecorder() // Reset the response recorder for each test case
			HandleRoutingError(ctx, mux, marshaler, w, r, tt.httpStatus)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
