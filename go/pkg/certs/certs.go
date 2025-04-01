// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"crypto/x509"
	"fmt"
	"os"
)

// GetCertPool loads the Certificate Authority from the given path
func GetCertPool(CaPath string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(CaPath)
	if err != nil {
		return nil, err
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append CA certificate from %s", CaPath)
	}
	return certPool, nil
}
