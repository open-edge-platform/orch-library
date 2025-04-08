<!---
  SPDX-FileCopyrightText: (C) 2024 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Golang Libraries for Orchestration Microservices

This repository includes common libraries that are used across orchestration microservices.

- **Middlewares**
  - Gin
    - Middleware for message size limiting
    - Middleware for checking non-Unicode characters in path and query parameters of REST API requests
    - Middleware for checking non-Unicode characters in the body of REST API requests
    - A custom routing error handler
- **Logging**
  - Loggers for the Gin Web Framework and K8s Controllers based on the
    [dazl](https://github.com/open-edge-platform/orch-library/tree/main/go/dazl) logging framework
- **Auto Generated REST Clients and Mocks**
  - Golang REST Client and mock for Open Policy Agent (OPA)
  - Golang REST Client and mock for OpenID Connect (OIDC)
- **Handling Files**
  - Utility functions to load OpenAPI specs and extract required information (e.g., paths)
- **Error Handling**
  - A typed error package for unifying error formats in Golang and utility functions to convert gRPC errors to typed errors
    and vice versa.
