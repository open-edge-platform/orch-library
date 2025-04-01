// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openidconnect

//go:generate oapi-codegen -generate client -old-config-style -package openidconnect -o oidc-client.go openapi.yaml
//go:generate oapi-codegen -generate types -old-config-style -package openidconnect -o oidc-types.go openapi.yaml
//go:generate mockgen -destination=oidc-mock-rest.go -package=openidconnect -source oidc-client.go ClientWithResponsesInterface
