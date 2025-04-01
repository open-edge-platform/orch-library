// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package openpolicyagent

//go:generate oapi-codegen -generate client -old-config-style -package openpolicyagent -o opa-client.go openapi.yaml
//go:generate oapi-codegen -generate types -old-config-style -package openpolicyagent -o opa-types.go openapi.yaml
//go:generate mockgen -destination=opa-mock-rest.go -package=openpolicyagent -source opa-client.go ClientWithResponsesInterface
