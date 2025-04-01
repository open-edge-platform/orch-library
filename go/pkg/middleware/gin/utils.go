// SPDX-FileCopyrightText: (C) 2022 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func HandleRoutingError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
	if httpStatus != http.StatusMethodNotAllowed {
		runtime.DefaultRoutingErrorHandler(ctx, mux, marshaler, w, r, httpStatus)
		return
	}
	// Use HTTPStatusError to customize the DefaultHTTPErrorHandler status code
	err := &runtime.HTTPStatusError{
		HTTPStatus: httpStatus,
		Err:        status.Error(codes.Unimplemented, http.StatusText(httpStatus)),
	}

	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}
