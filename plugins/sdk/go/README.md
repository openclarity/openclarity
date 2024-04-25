# Plugin SDK for Go

## Usage

Examples can be found in [example](example) dir. Module can be imported as:

```bash
go get "github.com/openclarity/vmclarity/plugins/sdk"
```

## Developer notes

All scanner plugins run as containers and expose scanning capabilities via [Scanner Plugin OpenAPI](../../openapi.yaml) REST server implementation.
Developers should ship their scanners as container images that run the REST server.

Configuration for the REST server can be found in [config.go](plugin/config.go).
