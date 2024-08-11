# Plugin SDK for Go

## Usage

```go
package main

import (
	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

// Your scanner plugin should implement required types.Scanner interface
var _ types.Scanner = &Scanner{}

type Scanner struct{}

func main() {
    plugin.Run(&Scanner{})
}
```

Check available [example](example) for a more complete implementation reference. 

## Developer notes

Plugins expose scanning capabilities via [Scanner Plugin OpenAPI](../openapi.yaml) REST server implementation.
Developers should ship their plugins as container images that run the REST server.
