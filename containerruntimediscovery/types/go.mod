module github.com/openclarity/openclarity/containerruntimediscovery/types

go 1.22.4

require github.com/openclarity/openclarity/api/types v0.7.2

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/getkin/kin-openapi v0.127.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oapi-codegen/oapi-codegen/v2 v2.3.1-0.20240819063527-fa048ac69eea // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/openclarity/openclarity/core v0.7.2 // indirect
	github.com/openclarity/openclarity/plugins/sdk-go v0.7.2 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/tools v0.21.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/openclarity/openclarity/api/types => ../../api/types
	github.com/openclarity/openclarity/core => ../../core
	github.com/openclarity/openclarity/plugins/sdk-go => ../../plugins/sdk-go
)
