# Plugin SDK for Python

## Usage

Examples can be found in [example](example) dir. Module can be installed via:

```bash
SDK_NAME="plugin"
SDK_PATH="plugins/sdk/python"
pip install -e "git+https://github.com/openclarity/vmclarity.git#egg=$SDK_NAME&subdirectory=$SDK_PATH" 
```

## Developer notes

All scanner plugins run as containers and expose scanning capabilities via [Scanner Plugin OpenAPI](../../openapi.yaml) REST server implementation.
Developers should ship their scanners as container images that run the REST server.

Configuration for the REST server can be found in [config.py](plugin/server/config.py).
