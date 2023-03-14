# How to Contribute

Thanks for your interest in contributing to KubeClarity! Here are a few general guidelines on contributing and
reporting bugs that we ask you to review. Following these guidelines helps to communicate that you respect the time of
the contributors managing and developing this open source project. In return, they should reciprocate that respect in
addressing your issue, assessing changes, and helping you finalize your pull requests. In that spirit of mutual respect,
we endeavor to review incoming issues and pull requests within 10 days, and will close any lingering issues or pull
requests after 60 days of inactivity.

Please note that all of your interactions in the project are subject to our [Code of Conduct](/CODE_OF_CONDUCT.md). This
includes creation of issues or pull requests, commenting on issues or pull requests, and extends to all interactions in
any real-time space e.g., Slack, Discord, etc.

## Table Of Contents

- [Reporting Issues](#reporting-issues)
- [Development](#development)
  - [Generating API code](#generating-api-code)
  - [Building KubeClarity Binaries](#building-kubeclarity-binaries)
  - [Building KubeClarity Containers](#building-kubeclarity-containers)
  - [Linting](#linting)
  - [Unit Tests](#unit-tests)
  - [Testing End to End](#testing-end-to-end)
- [Sending Pull Requests](#sending-pull-requests)
- [Other Ways to Contribute](#other-ways-to-contribute)

## Reporting Issues

Before reporting a new issue, please ensure that the issue was not already reported or fixed by searching through our
[issues list](https://github.com/openclarity/kubeclarity/issues).

When creating a new issue, please be sure to include a **title and clear description**, as much relevant information as
possible, and, if possible, a test case.

**If you discover a security bug, please do not report it through GitHub. Instead, please see security procedures in
[SECURITY.md](/SECURITY.md).**

## Development

### Building KubeClarity

`make build` will build all of the KubeClarity code and UI.

Makefile targets are provided to compile and build the KubeClarity binaries.
`make build-all-go` can be used to build all of the go components, but also
specific targets are provided, for example `make cli` and `make backend` to
build the specific components in isolation.

`make ui` is provided to just build the UI components.

### Building KubeClarity Containers

`make docker` can be used to build the KubeClarity containers for all of the
components. Specific targets for example `make docker-cli` and `make
docker-backend` are also provided.

`make push-docker` is also provided as a shortcut for building and then
publishing the KubeClarity containers to a registry. You can override the
destination registry like:

```
DOCKER_REGISTRY=docker.io/tehsmash make push-docker
```

You must be logged into the docker registry locally before using this target.

### Linting

`make lint` can be used to run the required linting rules over the code.
golangci-lint rules and config can be viewed in the `.golangcilint` file in the
root of the repo.

`make fix` is also provided which will resolve lint issues which are
automaticlly fixable for example format issues.

`make license` can be used to validate that all the files in the repo have the
correctly formatted license header.

### Unit tests

`make test` can be used run all the unit tests in the repo. Alternatively you
can use the standard go test CLI to run a specific package or test by going
into a specific modules directory and running:

```
cd cli
go test ./cmd/... -run <test name regex>
```

### Generating API code

After making changes to the API schema for example `api/swagger.yaml`, you can run `make
api` to regenerate the model, client and server code.

### Testing End to End

End to end tests will start and exercise a KubeClarity running on the local
container runtime. This can be used locally or in CI. These tests ensure that
more complex flows such as the CLI exporting results to the API work as
expected.

> ***Note***:
> If running Docker Desktop for Mac you will need to increase docker daemon
> memory to 8G. Careful, this will drain a lot from your computer cpu.

In order to run end-to-end tests locally:

```shell
# Build all docker images
make docker
# Replace Values In The KubeClarity Chart:
sed -i 's/latest/${{ github.sha }}/g' charts/kubeclarity/values.yaml
sed -i 's/Always/IfNotPresent/g' charts/kubeclarity/values.yaml
# Build the KubeClarity CLI
make cli
# Move the Built CLI into the E2E Test folder
mv ./cli/bin/cli ./e2e/kubeclarity-cli
# Run the end to end tests
make e2e
```

## Sending Pull Requests

Before sending a new pull request, take a look at existing pull requests and issues to see if the proposed change or fix
has been discussed in the past, or if the change was already implemented but not yet released.

We expect new pull requests to include tests for any affected behavior, and, as we follow semantic versioning, we may
reserve breaking changes until the next major version release.

## Other Ways to Contribute

We welcome anyone that wants to contribute to KubeClarity to triage and reply to open issues to help troubleshoot
and fix existing bugs. Here is what you can do:

- Help ensure that existing issues follows the recommendations from the _[Reporting Issues](#reporting-issues)_ section,
  providing feedback to the issue's author on what might be missing.
- Review and update the existing content of our [Wiki](https://github.com/openclarity/kubeclarity/wiki) with up-to-date
  instructions and code samples.
- Review existing pull requests, and testing patches against real existing applications that use KubeClarity.
- Write a test, or add a missing test case to an existing test.

Thanks again for your interest on contributing to KubeClarity!

:heart:
