# E2E tests

VMClarity's end-to-end tests use the [Ginkgo](https://onsi.github.io/ginkgo/) test framework and the [Gomega](https://onsi.github.io/gomega/) matcher/assertion library to leverage behavior driven testing (BDD).

This test module is composed by:

* a `testenv.go` file where the test environment is configured,
* a `docker-compose.override.yml` file which will merge and override the installation compose file to run end-to-end tests,
* a `suite_test.go` file where the steps to perform the setup and cleanup of the Ginkgo test suite are defined,
* a `helper.go` file where common methods are defined,
* and test case files.

## Run test

To run all end-to-end tests, use the following command:

```
go test -v -failfast -test.v -test.paniconexit0 -timeout 2h -ginkgo.v .
```

To run a particular end-to-end test file, use:

```
go test -v -failfast -test.v -test.paniconexit0 -timeout 2h -ginkgo.v --ginkgo.focus-file <go test file> .
```

## Write a new test

To add a new test, create a new `<test_name>_test.go` file in the current directory and use the following template:

```go
var _ = ginkgo.Describe("<add a brief test case description>", func() {
    reportFailedConfig := ReportFailedConfig{}
    ginkgo.Context("<describe conditions or inputs>", func() {
        ginkgo.It("<describe the behaviour or feature being tested>", func(ctx ginkgo.SpecContext) {
            <implement test code>
        })
    })
    ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
        if ginkgo.CurrentSpecReport().Failed() {
            reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
            ReportFailed(ctx, testEnv, client, &reportFailedConfig)
        }
    })
})
```

Additionally, check the available test cases (e.g. `basic_scan_test.go`) to get started.