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

- [Troubleshooting and Debugging](#troubleshooting-and-debugging)
- [Reporting Issues](#reporting-issues)
- [Development](#development)
  - [Generating API code](#generating-api-code)
  - [Building VMClarity Binaries](#building-vmclarity-binaries)
  - [Building VMClarity Containers](#building-vmclarity-containers)
  - [Linting](#linting)
  - [Unit Tests](#unit-tests)
  - [Testing End to End](#testing-end-to-end)
- [Sending Pull Requests](#sending-pull-requests)
- [Other Ways to Contribute](#other-ways-to-contribute)

## Troubleshooting and Debugging

Please see the troubleshooting and debugging guide [here](docs/troubleshooting.md).

## Reporting Issues

Before reporting a new issue, please ensure that the issue was not already reported or fixed by searching through our
[issues list](https://github.com/openclarity/vmclarity/issues).

When creating a new issue, please be sure to include a **title and clear description**, as much relevant information as
possible, and, if possible, a test case.

**If you discover a security bug, please do not report it through GitHub. Instead, please see security procedures in
[SECURITY.md](SECURITY.md).**

## Development

### Building VMClarity Binaries

Makefile targets are provided to compile and build the VMClarity binaries.
`make build` can be used to build all of the components, but also specific
targets are provided, for example `make build-cli` and `make build-backend` to
build the specific components in isolation.

### Building VMClarity Containers

`make docker` can be used to build the VMClarity containers for all of the
components. Specific targets for example `make docker-cli` and `make
docker-backend` are also provided.

`make push-docker` is also provided as a shortcut for building and then
publishing the VMClarity containers to a registry. You can override the
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

To lint the cloudformation template, `cfn-lint` can be used, see
https://github.com/aws-cloudformation/cfn-lint#install for instructions on how
to install it for your system.

### Unit tests

`make test` can be used run all the unit tests in the repo. Alternatively you
can use the standard go test CLI to run a specific package or test like:

```
go test ./cli/cmd/... -run Test_isSupportedFS
```

### Generating API code

After making changes to the API schema in `api/openapi.yaml`, you can run `make
api` to regenerate the model, client and server code.

### Testing End to End

`make e2e` can be used run the end-to-end tests in the repository.

For details on how to test VMClarity, please check the testing guide [here](docs/test_e2e.md) on how to perform a test on AWS and the instructions [here](e2e/README.md) on how to run and add new tests.

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
