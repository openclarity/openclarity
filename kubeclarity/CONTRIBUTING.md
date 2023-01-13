# Welcome to the KubeClarity Contributing Guide

Thank you for your interest in contributing to KubeClarity!

The following is a set of guidelines and instructions to help you:

- Submit a bug report or create a feature request
- Build and test the KubeClarity code
- Make a code change and submit a pull request

If anything doesn't make sense or doesn't work when you run it please open an
issue and let us know!

## Table of Contents

- [Getting Started](#getting-started)
- [Code of Conduct](#code-of-conduct)
- [Creating Issues](#creating-issues)
  - [Submit a Bug Report](#submit-a-bug-report)
  - [Submit a Feature Request](#submit-a-feature-request)
- [Building and Testing KubeClarity](#building-and-testing-kubeclarity)
  - [Building KubeClarity](#building-kubeclarity)
  - [Unit Tests](#unit-tests)
  - [End to End Tests](#end-to-end-tests)
- [Making Changes](#making-changes)
  - [Exercising your change](#exercising-your-change)
  - [Submitting a Change For Review](#submitting-a-change-for-review)

## Getting Started

TODO

## Code of Conduct

TODO

## Creating Issues

TODO

### Submit a Bug Report

TODO

### Submit a Feature Request

TODO

## Building and Testing KubeClarity

TODO

### Building KubeClarity

TODO

### Unit Tests

To run all the unit tests locally a make target is provided:

```shell
make test
```

To run an individual test or test suite directly then you can go into the
component directory and use `go test`:

```shell
cd cli
go test ./...
```

### End to End Tests

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
sed -i 's/latest/v1.1/g' charts/kubeclarity/Chart.yaml
sed -i 's/latest/${{ github.sha }}/g' charts/kubeclarity/values.yaml
sed -i 's/Always/IfNotPresent/g' charts/kubeclarity/values.yaml
# Build the KubeClarity CLI
make cli
# Move the Built CLI into the E2E Test folder
mv ./cli/bin/cli ./e2e/kubeclarity-cli
# Run the end to end tests
make e2e
```

## Making Changes

TODO

### Exercising a change

While making a change to KubeClarity its worth testing locally to ensure that
no regressions have been introduced and to ensure that everything is still
working as expected. Its also good practise to test locally before submitting
the code for review to prevent extra iterations on the PR due to failing CI
testing. Please refer to the
[Building and Testing KubeClarity](#building-and-testing-kubeclarity) section
for more details.

### Submitting a Change for Review

TODO
