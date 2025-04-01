// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMessageSizeLimiter tests message size limiter middleware function that will be used in the REST server
func TestMessageSizeLimiter(t *testing.T) {
	const maxMessageSize = 1024 * 1024
	testCases := []struct {
		name           string
		messageSize    int
		expectedStatus int
	}{
		{
			name:           "Message with half of the size limit",
			messageSize:    maxMessageSize / 2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Message exceeds size limit",
			messageSize:    maxMessageSize + 1,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Message with the size limit",
			messageSize:    maxMessageSize,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(MessageSizeLimiter(maxMessageSize))
			router.POST("/test", func(_ *gin.Context) {

			})

			// Create a test server
			server := httptest.NewServer(router)
			defer server.Close()

			// Prepare the request body
			body := bytes.NewReader(make([]byte, tc.messageSize))

			// Send a test request to the server
			req, err := http.NewRequest(http.MethodPost, server.URL+"/test", body)
			assert.NoError(t, err, "Failed to create the request")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err, "Failed to send the request")

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				assert.NoError(t, err)
			}(resp.Body)

			// Check if the response status matches the expected status
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Unexpected status code")

			// If the request was rejected due to size, check the error message
			if resp.StatusCode == http.StatusRequestEntityTooLarge {
				respBody, _ := io.ReadAll(resp.Body)
				expectedErrMsg := fmt.Sprintf("Message size exceeds the limit of %d bytes", maxMessageSize)
				assert.Contains(t, string(respBody), expectedErrMsg, "Expected error message not found")

			}
		})
	}
}

func TestUnicodeChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(UnicodePrintableCharsChecker())
	router.POST("/test", func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		c.String(http.StatusOK, string(bodyBytes))
	})

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()
	testCases := []struct {
		body           string
		expectedStatus int
	}{
		{"This is valid string body message", http.StatusOK},
		{"This message body has a null char \x00", http.StatusBadRequest},
		{"This message body has a non printable character: \x1F", http.StatusBadRequest},
		{"This message body has a delete character as a non printable character: \x7F", http.StatusBadRequest},
		{"This message body has new line character as an acceptable character \n", http.StatusOK},
		{"This message body has carriage return as an acceptable character \r", http.StatusOK},
	}
	for _, tc := range testCases {
		t.Run(tc.body, func(t *testing.T) {
			// Prepare the request body
			body := strings.NewReader(tc.body)
			// Send a test request to the server
			req, err := http.NewRequest(http.MethodPost, server.URL+"/test", body)
			assert.NoError(t, err, "Failed to create the request")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err, "Failed to send the request")

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				assert.NoError(t, err)
			}(resp.Body)
			// Check if the response status matches the expected status
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Unexpected status code")

			// Check the body
			if resp.StatusCode == http.StatusOK {
				bodyReq := req.GetBody
				bodyReader, err := bodyReq()
				assert.NoError(t, err)
				bodyBytes, err := io.ReadAll(bodyReader)
				assert.NoError(t, err)

				respBodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, string(bodyBytes), string(respBodyBytes))
			}

		})
	}
}

func TestPathParamUnicodeChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PathParamUnicodeCheckerMiddleware())
	router.GET("/test/:param", func(_ *gin.Context) {})

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()
	testCases := []struct {
		name            string
		path            string
		expectedErrCode int
	}{
		{
			name:            "Test with Invalid UTF8 path parameter",
			path:            "/test/%80",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with valid UTF8 path parameter",
			path:            "/test/valid",
			expectedErrCode: http.StatusOK,
		},
		{
			name:            "Test with invalid query parameter (case #1)",
			path:            "/test/valid?label=n@U8dm[5h+l|4pxD.%ad&wYT",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #2)",
			path:            "/test/valid?label=)ndWYD%c41k\\\"DjM.*",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #3)",
			path:            "/test/valid?label=~ne`X\\nfecPD1k8{Vnx%bB!uB't/@)_qVvb0eL)UZt\"d",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #4)",
			path:            "/test/valid?label=)E`lOE0:r`_6w%23nY'DJu*Gd6^YbBJQ!P%BASXzxTI\\a",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #5)",
			path:            "/test/valid?label=^%d3Xt&n,\"z5",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #6)",
			path:            "/test/valid?label=GIW+:_~O)%8c^/H&o",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #7)",
			path:            "/test/valid?label=Fp5y_G`c%A5&!yQqR=80d",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #8)",
			path:            "/test/valid?label=e@S)ilN=Lsi|4_vh(5L9Rg)%23j9_%a2v>8.kStPJg=EKUr",
			expectedErrCode: http.StatusBadRequest,
		},
		{
			name:            "Test with invalid query parameter (case #9)",
			path:            "/test/valid?label=S2BjE\"k_`Wgv6S%a72Iz1O*\"",
			expectedErrCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, server.URL+tc.path, nil)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedErrCode, resp.StatusCode)

		})
	}
}
