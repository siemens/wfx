<p align="center" width="100%"><img src="hugo/static/images/logo.svg" width="400"/></p>

# wfx

A lightweight, general-purpose workflow executor.

[![CI](https://github.com/siemens/wfx/actions/workflows/ci.yml/badge.svg)](https://github.com/siemens/wfx/actions/workflows/ci.yml)
[![Pages](https://github.com/siemens/wfx/actions/workflows/pages.yml/badge.svg)](https://github.com/siemens/wfx/actions/workflows/pages.yml)
[![Coverage](https://codecov.io/github/siemens/wfx/coverage.svg?branch=main)](https://codecov.io/github/siemens/wfx?branch=main)

## Overview

[_Workflows_](docs/workflows.md) are modeled as finite-state machines and are instantiated as [_Jobs_](docs/workflows.md#jobs) through which the wfx and a client progress in lock-step to perform a task.
Such a task could be [software updating](workflow/dau/README.md) the client, progressing a work item through [Kanban](docs/workflows.md#hands-on-playing-kanban), … in essence anything requiring cooperation and coordination.

As being general purpose, wfx is only concerned with driving the (state) machinery, the specific actions a client should perform are encoded in the client implementation(s).
Hence, one wfx instance can drive a multitude of wholly different workflows.
Instantiating a workflow as a job augments it with additional metadata, the [_Job Definition_](docs/workflows.md#jobs), which contains job-specific information such as, e.g., URLs or other data the client (implementation) can make use of for this particular job.

To illustrate the concepts as well as the wfx / client interaction, consider the following figure

```txt
┌──────────────────────────────────────┐                       ┌──────────────┐
│                 wfx                  │                     ┌─┴────────────┐ │
│                                      │                     │   Client Y   │ │
│                                      │     poll for jobs   │              │ │
│                instantiate ┌───────┐ │◀────────────────────│              │ │
│  ┌────────────┐        ┌──▶│ Job 1 │ │────────────────────▶│──────┐       │ │
│  │ Workflow A ├────────┤   └───────┘ │◀─┐  job information │      ▼       │ │
│  └────────────┘        │   ┌───────┐ │  │                  │     act      │ │
│                   ┌────┼──▶│ Job 2 │ │  └──────────────────│◀─────┘       │ │
│  ┌────────────┐   │    │   └───────┘ │     update state    │              │ │
│  │ Workflow B ├───┘    │   ┌───────┐ │                     │              │ │
│  └────────────┘        └──▶│ Job 3 │ │           .         │              │ │
│                            └───────┘ │           .         │              │ │
│       ...                     ...    │           .         │              │ │
│                                      │                     │              ├─┘
└──────────────────────────────────────┘                     └──────────────┘
```

with the wfx having loaded a number of workflows `Workflow A`, `Workflow B`, … that got instantiated as `Job 1`, `Job 2`, `Job 3`, … with a `Client Y` working on `Job 1`: It polls the wfx for a new job or the current job's status, in return receives the job information, performs actions if applicable, and finally reports the new job status back to the wfx. This lock-step procedure is repeated until the workflow reaches a terminal state which could be identified with, e.g., success or failure.

**wfx in (Example) Action**

An exemplary [Kanban](https://en.wikipedia.org/wiki/Kanban)-inspired [workflow](docs/workflows.md#kanban-example-workflow) illustrating the interplay between the wfx as Kanban "Board", a Product Owner creating jobs, and a Developer executing them:

![Konsole Demo](share/demo/kanban/demo.gif)

**wfx Features & Non-Features**

- Design Guidelines
  - Compact, scalable core focusing on the essentials
  - Proper interfaces to external systems for modularity and integrability:
    Accompanying and necessary services like artifact storage and device registry
    are likely already available or are better provided by specialized solutions
- Implementation
  - Extendable modularized source code architecture
  - Lightweight, no dependencies (statically linked binaries)
  - Efficient, native code for a wide variety of platforms and operating systems (as supported by the [Go](https://golang.org/) Language)
  - Fully documented REST API, see [wfx OpenAPI Specification](spec/wfx.openapiv3.yml)
  - Extensive test suite including load tests
- Deployment / Usability
  - Load / Unload workflows at run-time
  - Hot / Live reload of configuration file
  - Persistent Storage: built-in support for [SQLite](https://www.sqlite.org/), [PostgreSQL](https://www.postgresql.org) and [MySQL](https://www.mysql.com)
  - A complimentary built-in file server serving as artifact storage for dynamic deployments and integration without external file storage
  - Transport Layer Security (HTTPS) with support for custom certificates

**wfx Clients**

Currently, the following clients have support for wfx:

- [SWUpdate](https://github.com/sbabic/swupdate) - Software Update for Embedded Linux Devices implementing support for the [Device Artifact Update Workflow Family](workflow/dau/README.md)

## Documentation

Grouped by topic, the following documentation is available in [docs/](docs/):

- [Workflows](docs/workflows.md): Core concepts of Workflows and Jobs
- [Installation](docs/installation.md): How to install and deploy wfx.
- [Configuration](docs/configuration.md): How to configure wfx.
- [Operation](docs/operations.md): How to operate wfx.
- [Use-Cases](docs/use-cases.md): A collection of use-cases for wfx.
- [API](docs/operations.md#api): wfx's north- and southbound APIs.

You can also browse the rendered documentation at <https://siemens.github.io/wfx/>.

## Roadmap

The roadmap is tracked via [Github issues](https://github.com/siemens/wfx/issues).

## Contributing

Contributions are encouraged and welcome!

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

Copyright ©️ 2023 Siemens AG.

Released under the [Apache-2.0](LICENSE) license.
