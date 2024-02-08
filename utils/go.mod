module github.com/openclarity/vmclarity/utils

go 1.21.4

require (
	github.com/mitchellh/mapstructure v1.5.0
	github.com/moby/sys/mountinfo v0.7.1
	github.com/onsi/gomega v1.31.1
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// NOTE(akijakya): replace is required for the following issue: https://github.com/mitchellh/mapstructure/issues/327,
// which has been solved in the go-viper fork.
// Remove replace if all packages using the original repo has been switched to this fork (or at least viper:
// https://github.com/spf13/viper/pull/1723)
replace github.com/mitchellh/mapstructure => github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1
