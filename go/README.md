<!---
  SPDX-FileCopyrightText: (C) 2024 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Golang Libraries for Orchestration Microservices

This repo includes common libraries that are used across orchestration microservices.

- **Middlewares**
  - Gin
    - middleware for message size limiting
    - middleware for checking non unicode characters in path and query parameters of REST API requests
    - middleware for checking non unicode characters in body of REST API requests
    - A custom routing error handler
- **Logging**
  - Loggers for Gin Web Framework and K8s Controllers based on
    [dazl](https://github.com/open-edge-platform/orch-library/go/dazl) logging framework
- **Auto Generated REST Clients and Mocks**
  - Golang REST Client and mock for Open Policy Agent (OPA)
  - Golang REST Client and mock for Open ID Connect (OIDC)
- **Handling Files**
  - Utility functions to load OpenAPI specs and extract required info (e.g. paths)
- **Error Handling**
  - A Typed error package for unifying error format in Golang and utility functions to convert gRPC errors to typed errors
    and vice versa.
