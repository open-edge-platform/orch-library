// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"unicode"
	"unicode/utf8"
)

// MessageSizeLimiter a middleware to reject requests that have content length greater than maxMessageSize
func MessageSizeLimiter(maxMessageSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentLength := c.Request.ContentLength
		if contentLength < 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "The request is not valid, message size is negative",
				"details": []string{},
			})
			return
		}

		// Check if the incoming message size is greater than the allowed maxMessageSize
		if contentLength > maxMessageSize {
			// If the message is too large, return an error
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    http.StatusRequestEntityTooLarge,
				"message": fmt.Sprintf("Message size exceeds the limit of %d bytes", maxMessageSize),
				"details": []string{},
			})
			return
		}
		// If the message size is within the limit, continue processing the request
		c.Next()
	}
}

// UnicodePrintableCharsChecker checks if the request body contains just unicode characters and returns error if it
// finds any non unicode characters
func UnicodePrintableCharsChecker() gin.HandlerFunc {
	return func(c *gin.Context) {

		bodyCopy := new(bytes.Buffer)
		_, err := io.Copy(bodyCopy, c.Request.Body)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		currentReader := bytes.NewReader(bodyCopy.Bytes())
		nextReader := io.NopCloser(bytes.NewReader(bodyCopy.Bytes()))

		for {
			r, _, err := currentReader.ReadRune()
			if err != nil {
				if err == io.EOF {
					break
				}
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "Request body contains non printable characters",
					"details": []string{},
				})
				return
			}
		}

		c.Request.Body = nextReader
		c.Next()

	}
}

func PathParamUnicodeCheckerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, v := range c.Request.URL.Query() {
			for _, queryParam := range v {
				if !utf8.ValidString(queryParam) {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusBadRequest,
						"message": "Invalid UTF8 query parameter",
						"details": []string{},
					})
					c.Abort()
					return

				}
			}
		}

		for _, param := range c.Params {
			if !utf8.ValidString(param.Value) {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": "Invalid UTF8 path parameter",
					"details": []string{},
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
