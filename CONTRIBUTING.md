# How to Contribute

Thanks for your interest in contributing to VMClarity! Here are a few general guidelines on contributing and
reporting bugs that we ask you to review. Following these guidelines helps to communicate that you respect the time of
the contributors managing and developing this open source project. In return, they should reciprocate that respect in
addressing your issue, assessing changes, and helping you finalize your pull requests. In that spirit of mutual respect,
we endeavor to review incoming issues and pull requests within 10 days, and will close any lingering issues or pull
requests after 60 days of inactivity.

Please note that all of your interactions in the project are subject to our [Code of Conduct](CODE_OF_CONDUCT.md). This
includes creation of issues or pull requests, commenting on issues or pull requests, and extends to all interactions in
any real-time space e.g., Slack, Discord, etc.

## Table Of Contents

- [Reporting Issues](#reporting-issues)
- [Development](#development)
  - [Dependencies](#dependencies)
  - [Development Environment](#development-environment)
  - [Running the VMClarity stack locally using Docker](#running-the-vmclarity-stack-locally-using-docker)
  - [Building VMClarity Binaries](#building-vmclarity-binaries)
  - [Building VMClarity Containers](#building-vmclarity-containers)
  - [Linting](#linting)
  - [Unit Tests](#unit-tests)
  - [Testing End to End](#testing-end-to-end)
  - [Troubleshooting and Debugging](#troubleshooting-and-debugging)
- [Sending Pull Requests](#sending-pull-requests)
- [Other Ways to Contribute](#other-ways-to-contribute)

## Reporting Issues

Before reporting a new issue, please ensure that the issue was not already reported or fixed by searching through our
[issues list](https://github.com/openclarity/vmclarity/issues).

When creating a new issue, please be sure to include a **title and clear description**, as much relevant information as
possible, and, if possible, a test case.

**If you discover a security bug, please do not report it through GitHub. Instead, please see security procedures in
[SECURITY.md](SECURITY.md).**

## Development

After cloning the repository, you can run `make help` to inspect the targets that are used for checking, generating,
building, and publishing code.

### Dependencies

- `Docker` (for local and e2e testing)
- `Go` for the backend (the current version used by the project can be found in the .go-version file)
- A Node package manager for the frontend, such as `npm` or `yarn`

Internal dependencies by `make` targets are automatically installed if not present locally.

### Development Environment

Depending on your IDE/editor of choice, you might need a `go.work` file for the `gopls` language server to find all
references properly, such as:

```text
go 1.22.4

use (
  ./api/client
  ./api/server
  ./api/types
  ./cli
  ./containerruntimediscovery/client
  ./containerruntimediscovery/server
  ./containerruntimediscovery/types
  ./core
  ./e2e
  ./e2e/testdata
  ./installation
  ./orchestrator
  ./provider
  ./plugins/runner
  ./plugins/sdk-go
  ./plugins/sdk-go/example
  ./plugins/store/kics
  ./scanner
  ./testenv
  ./uibackend/client
  ./uibackend/server
  ./uibackend/types
  ./utils
    ./workflow)
```

### Running the VMClarity stack locally using Docker

For testing the changes across the whole stack, VMClarity can be ran with Docker provider locally, after the images have
been [built](#building-vmclarity-containers) and their tags have been updated in the
`installation/docker/image_override.env` file:

```shell
docker compose --project-name vmclarity \
               --file installation/docker/docker-compose.yml \
               --env-file installation/docker/image_override.env \
               up -d --wait --remove-orphans
```

When working only on one stack component, the component in question can be commented out in the [docker compose
file](installation/docker/docker-compose.yml) and ran separately with `go run`, or in the case of the UI, with the
following commands:

```shell
cd ui
npm install
npm start
```

Update the [NGINX config](installation/docker/gateway.conf) accordingly if the components in question are affected to
ensure that Docker can communicate with them if they are ran on local network.
Some environment variables could also be necessary for you to export in your shell before running the component, inspect
the contents of the corresponding `.env` file in the `installation/docker` directory!

To clean up the VMClarity stack locally, run:

```shell
docker compose --project-name vmclarity \
               --file installation/docker/docker-compose.yml \
               down --remove-orphans --volumes
```

### Building VMClarity Binaries

Makefile targets are provided to compile and build the VMClarity binaries. `make build` can be used to build all the
components, while `make build-all-go` and `make ui` only builds the go modules or the UI.

### Building VMClarity Containers

`make docker` can be used to build the VMClarity containers for all the components. Specific targets for example `make
docker-cli` and `make docker-ui-backend` are also provided.

In order to also publish the VMClarity containers to a registry, please set the `DOCKER_PUSH` environment variable to
`true`. You can override the destination registry as well:

```shell
DOCKER_REGISTRY=docker.io/my-vmclarity-images DOCKER_PUSH=true make docker
```

You must be logged into the docker registry locally before using this target.

### Linting

`make lint` can be used to run all the required linting rules over the code. In this case, the following targets will be
ran:

- `make license-check` can be used to validate that all the files in the repo have the correctly formatted license
header.
- `make lint-actions` checks Github Actions workflow files.
- `make lint-bicep` lints Bicep files.
- `make lint-cfn` lints Cloudformation files.
- `make lint-go` runs `golangci-lint` on the Go files. Rules and config can be viewed in the `.golangci.yml` file in the
root of the repo.
- `make lint-helm` lints the Helm chart.

`make fix` is also provided which can automatically resolve lint issues such as formatting.

### Unit tests

`make test` can be used run all the unit tests in the repo. Alternatively you can use the standard go test CLI to run a
specific package or test like:

```shell
go test ./cli/cmd/... -run Test_isSupportedFS
```

### Generators

`make gen` runs the following targets that can be ran separately as well:

- After making changes to the API schema in `api/openapi.yaml`, you can run `make gen-api` to regenerate the model,
client and server code.
- Run `make gen-bicep` for generating bicep files after modifying them for installing VMClarity on Azure.
- Run `make gen-helm-docs` for generating the docs after making changes to VMClarity's Helm chart.

### Testing End to End

`make e2e-docker` can be used run the end-to-end tests in the repository locally using Docker. `make e2e-k8s` can also
be used to run end-to-end tests for Kubernetes provider using Docker.

For details on how to test VMClarity, please check the testing guide [here](docs/test_e2e.md) on how to perform a test
on AWS and the instructions [here](e2e/README.md) on how to run and add new tests.

### Troubleshooting and Debugging

Please see the troubleshooting and debugging guide [here](docs/troubleshooting.md).

## Sending Pull Requests

Before sending a new pull request, take a look at existing pull requests and issues to see if the proposed change or fix
has been discussed in the past, or if the change was already implemented but not yet released.

We expect new pull requests to include tests for any affected behavior, and, as we follow semantic versioning, we may
reserve breaking changes until the next major version release.

## Other Ways to Contribute

We welcome anyone that wants to contribute to VMClarity to triage and reply to open issues to help troubleshoot
and fix existing bugs. Here is what you can do:

- Help ensure that existing issues follows the recommendations from the _[Reporting Issues](#reporting-issues)_ section,
  providing feedback to the issue's author on what might be missing.
- Review and update the existing content of our [Wiki](https://github.com/openclarity/vmclarity/wiki) with up-to-date
  instructions and code samples.
- Review existing pull requests, and testing patches against real existing applications that use VMClarity.
- Write a test, or add a missing test case to an existing test.

Thanks again for your interest on contributing to VMClarity!

:heart:
