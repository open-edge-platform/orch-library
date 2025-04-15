<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Orchestration Library

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

This library is a collection of Go language packages that are used to build the Orchestrator.

The library includes common functionality that is used across many of the components of the Orchestrator, such as:
[Application Catalog], [App Deployment Manager], [Tenant Controller], and [Cluster Manager].

Please see the [go/pkg](go/pkg) directory for the list of packages that are included in the library, and see the
[go/README](go/README.md) document for more details.

## Get Started

To include the library in your project, import the required package as a Go module into your Go code:

```text
import "github.com/open-edge-platform/orch-library/go/pkg/auth"
```

## Develop

To add a new package to the library, create a new directory in the `go/pkg` directory and add the Go code to the
directory. Add unit tests for your new Go code. Do not duplicate code that is already in the library.

If you wish to enhance an existing package, please open a pull request with your changes.

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed
and merged into the `main` branch. We encourage you to add appropriate unit tests and e2e tests if your contribution introduces
a new feature. See the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.

Additionally, ensure the following commands are successful:

```shell
make test
make lint
make license
```

## Community and Support

To learn more about the project, its community, and governance, visit the Edge Orchestrator Community.
For support, start with Troubleshooting or contact us.

## License

Orchestration Library is licensed under [Apache 2.0 License](LICENSES/Apache-2.0.txt).

[Application Catalog]: https://github.com/open-edge-platform/app-orch-catalog
[App Deployment Manager]: https://github.com/open-edge-platform/app-orch-deployment
[Cluster Manager]: https://github.com/open-edge-platform/cluster-manager
[Tenant Controller]: https://github.com/open-edge-platform/app-orch-tenant-controller
