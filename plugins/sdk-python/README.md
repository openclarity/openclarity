# Plugin SDK for Python

## Usage

```python
# Plugin package installed from:
# git+https://github.com/openclarity/vmclarity#egg=plugin&subdirectory=plugins/sdk-python

from plugin.scanner import AbstractScanner
from plugin.server import Server


# Your scanner plugin should implement required AbstractScanner interface
class ExampleScanner(AbstractScanner):
    def __init__(self):
        return


if __name__ == '__main__':
    Server.run(ExampleScanner())
```

Check available [example](example) for a more complete implementation reference.

## Developer notes

Plugins expose scanning capabilities via [Scanner Plugin OpenAPI](../openapi.yaml) REST server implementation.
Developers should ship their plugins as container images that run the REST server.
