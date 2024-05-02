# Plugin SDK for Python

## Usage

Examples can be found in [example](example) dir. Module can be installed via:

```bash
SDK_VERSION="main" # use latest package from "main" branch
# SDK_VERSION="v0.7.0" # use specific git release tag
pip install -e "git+https://github.com/openclarity/vmclarity@$SDK_VERSION#egg=plugin&subdirectory=plugins/sdk/python" 
```

## Developer notes

All scanner plugins run as containers and expose scanning capabilities via [Scanner Plugin OpenAPI](../../openapi.yaml) REST server implementation.
Developers should ship their scanners as container images that run the REST server.

Configuration for the REST server can be found in [config.py](plugin/server/config.py).

## TODO

- Add testing logic to verify that SDK works
- Consider using [Rye](https://rye-up.com/) for Python package management
