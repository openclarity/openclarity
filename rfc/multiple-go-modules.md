# [RFC] Split VMClarity into multiple go modules

*Note: this RFC template follows HashiCrop RFC format described [here](https://works.hashicorp.com/articles/rfc-template)*

|               |                                                               |
|---------------|---------------------------------------------------------------|
| **Created**   | 2024-01-17                                                    |
| **Status**    | WIP \| InReview \| **Approved** \| Obsolete                   |
| **Owner**     | *paralta*                                                     |
| **Approvers** | [PR-1105](https://github.com/openclarity/vmclarity/pull/1105) |

---

This RFC proposes a new project structure which splits the main module into multiple modules to allow us to version modules separately and reduce dependencies.

## Background

VMClarity consists of a single module containing all the packages required to run VMClarity. However, there is a clear boundary between the following packages: api, cli, orchestrator, ui and uibackend.

The build and deployment of these packages is already performed separately and independently.

## Proposal

The proposal here is to split the VMClarity repository into multiple modules:

- **api**. Interface between all the services in VMClarity including the DB. Composed by API model, backend client and server.
- **cli**. Responsible for running a scan in an asset and report the results back to api. Contains the logic to configure, run and manage different analysers and scanners. 
- **orchestrator**. Responsible for managing scan configurations, scans, assets and estimations.
- **provider**. Responsible for discovery and scan infrastructure setup for each provider. Contains logic to find assets and run scans on AWS, GCP, Azure, Docker and Kubernetes.
- **uibackend**. Responsible for offloading the ui from data processing and filtering. Slightly coupled with ui. Composed by API model, backend client and server.
- **utils**. Contains packages shared between modules.

Each module will have its own go.mod file and each module will be versioned independently.

## Implementation

The scope of this RFC is not to change code logic but to change code structure. Therefore, the following table describes the path changes for each package impacted.

| Module           | Current path                  | New path                         | Version Tag                             |
| ---------------- | ----------------------------- | -------------------------------- | --------------------------------------- |
| api              | api                           | api                              | api/v0.7.0                              |
| api/server       | pkg/apiserver                 | api/server                       | api/server/v0.7.0                       |
| api/client       | pkg/shared/backendclient      | api/client                       | api/client/v0.7.0                       |
| cli              | pkg/cli                       | cli                              | cli/v0.7.0                              |
| cli              | pkg/shared/analyzer           | cli/pkg/analyzer                 | cli/v0.7.0                              |
| cli              | pkg/shared/config             | cli/pkg/config                   | cli/v0.7.0                              |
| cli              | pkg/shared/converter          | cli/pkg/converter                | cli/v0.7.0                              |
| cli              | pkg/shared/families           | cli/pkg/families                 | cli/v0.7.0                              |
| cli              | pkg/shared/findingkey         | cli/pkg/findingkey               | cli/v0.7.0                              |
| cli              | pkg/shared/job_manager        | cli/pkg/jobmanager               | cli/v0.7.0                              |
| cli              | pkg/shared/scanner            | cli/pkg/scanner                  | cli/v0.7.0                              |
| cli              | pkg/shared/utils              | cli/pkg/utils                    | cli/v0.7.0                              |
| orchestrator     | pkg/orchestrator              | orchestrator                     | orchestrator/v0.7.0                     |
| uibackend        | pkg/uibackend                 | uibackend                        | uibackend/v0.7.0                        |
| uibackend/server | pkg/uibackend/rest            | uibackend/server                 | uibackend/server/v0.7.0                 |
| uibackend/client | pkg/shared/uibackendclient    | uibackend/client                 | uibackend/client/v0.7.0                 |
| utils            | pkg/version                   | utils/version                    | utils/v0.7.0                            |
| utils            | pkg/shared/command            | utils/command                    | utils/v0.7.0                            |
| utils            | pkg/shared/fsutils            | utils/fsutils                    | utils/v0.7.0                            |
| utils            | pkg/shared/log                | utils/log                        | utils/v0.7.0                            |
| utils            | pkg/shared/manifest           | utils/manifest                   | utils/v0.7.0                            |
| provider         | pkg/orchestrator/provider     | provider                         | provider/v0.7.0                         |
| provider         | pkg/containerruntimediscovery | provider/common/runtimediscovery | provider/common/runtimediscovery/v0.7.0 |
| e2e              | e2e                           | e2e                              | e2e/v0.7.0                              |
| testenv          | e2e/testenv                   | testenv                          | testenv/v0.7.0                          |

To improve compliance with https://github.com/golang-standards/project-layout, the changes below are also proposed.

| Module       | Current path                     | New path                      |
| ------------ | -------------------------------- | ----------------------------- |
| provider     | example_external_provider_plugin | provider/examples/external    |
| cli          | scanner_boot_test                | cli/test/boot                 |
|              | img                              | assets                        |

Makefile, GitHub workflows and other files will need to be updated to point to the new paths.

The VMClarity directory will have the following structure.

```sh
.
├── Makefile
├── api
│   ├── client
│   │   ├── client.cfg.yaml
│   │   └── go.mod
│   ├── go.mod
│   ├── models
│   │   └── models.cfg.yaml
│   ├── openapi.yaml
│   └── server
│       ├── cmd
│       ├── go.mod
│       ├── pkg
│       │   ├── common
│       │   ├── database
│       │   └── rest
│       └── server.cfg.yaml
├── cli
│   ├── cmd
│   ├── go.mod
│   ├── pkg
│   │   ├── analyzer
│   │   ├── config
│   │   ├── converter
│   │   ├── families
│   │   ├── findingkey
│   │   ├── jobmanager
│   │   ├── presenter
│   │   ├── scanner
│   │   ├── state
│   │   └── utils
│   └── test
│       └── boot
├── docs
├── e2e
│   ├── config
│   └── go.mod
│── testenv
│   ├── docker
│   ├── kubernetes
│   ├── types
│   ├── utils
│   └── go.mod
├── assets
├── installation
│   ├── aws
│   ├── azure
│   ├── docker
│   ├── gcp
│   └── kubernetes
├── orchestrator
│   ├── cmd
│   ├── go.mod
│   └── pkg
│       ├── assetscanestimationwatcher
│       ├── assetscanprocessor
│       ├── assetscanwatcher
│       ├── common
│       ├── discovery
│       ├── scanconfigwatcher
│       ├── scanestimationwatcher
│       └── scanwatcher
├── provider
│   ├── cmd
│   ├── examples
│   │   └── external
│   ├── go.mod
│   └── pkg
│       ├── aws
│       ├── azure
│       ├── cloudinit
│       ├── common
│           └── containerruntimediscovery
│               └── cmd
│       ├── docker
│       ├── external
│       ├── gcp
│       └── kubernetes
├── rfc
├── ui
│   └── src
├── uibackend
│   ├── client
│   │   └── client.cfg.yaml
│   ├── go.mod
│   ├── models
│   │   └── models.cfg.yaml
│   ├── openapi.yaml
│   └── server
│       ├── cmd
│       ├── go.mod
│       ├── pkg
│       └── server.cfg.yaml
└── utils
    ├── command
    ├── fsutils
    ├── go.mod
    ├── log
    ├── manifest
    └── version
```

# Release

Each module will have a tag with the format `prefix/version` where prefix is the directory within the repository where the module is defined, more details [here](https://go.dev/wiki/Modules#publishing-a-release). For now, the same version will be used for each module even if there are no changes, this will simplify managing compatibility between modules across versions.

The release process will be updated to cope with tagging multiple modules. Not all steps will be automated with GitHub actions some will require new scripts in `Makefile` so a new `docs/release.md` file will be created with instructions on how to perform a release.

Example of release instructions for version 0.7.0:

1. [New Makefile Script] Create pull request with version bumps for all modules in repository. E.g. to bump the api module, the require section in `go.mod` files should be updated like
    ```
     - github.com/openclarity/vmclarity/api v0.0.0
     + github.com/openclarity/vmclarity/api v0.7.0
    ```
2. [New Makefile Script] Create and push tags with v0.7.0 that points to the commit performed in the previous step. E.g. create tag for the api module with
    ```
    git tag -a api/v0.7.0
    ```
3. [GitHub Action] Current release process
4. [Extend GitHub Action] Once a stable release is successfully published, each module will be tagged. E.g. the api module will be tagged with the tag created in step 2, `api/v0.7.0`.

## UX/UI

This RFC has no user-impacting changes.
