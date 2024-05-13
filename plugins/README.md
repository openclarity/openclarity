# Overview

**Plugins** provide additional **scanning capabilities** to VMClarity ecosystem.
Project structure:

- **runner** - Provides necessary logic to manage scanner plugins in VMClarity.
- **sdk-*** - Language-specific libraries, templates, and examples to aid with the implementation of scanner plugins.
- **store** - Collection of available plugins that can be directly used in VMClarity.

### Requirements

Scanner plugins are distributed as containers and require [**Docker Engine**](https://docs.docker.com/engine/) on the host that runs the actual scanning via
VMClarity CLI to work.

## Support

✅ List of supported environments:
1. AWS
2. GCP
3. Azure
4. Docker

❌ List of unsupported environments:
- _Kubernetes_ - We plan on adding plugin support to Kubernetes once we have dealt with all the security considerations.

_Note:_ Plugin support has been tested against [VMClarity installation artifacts](../installation) for the given environments.

## Usage

You can start using plugins via **[Plugins Store](store)**.
For example, you can pass the `.families.yaml` scan config file defined below to the VMClarity CLI `scan` command.
This configuration uses **KICS scanner** to scan `/tmp` dir for IaC security misconfigurations.

```yaml
# --- .families.yaml
plugins:
  enabled: true
  scanners_list:
    - "kics"
  inputs: 
    - input: "/tmp"
      input_type: "rootfs"
  scanners_config:
    kics:
      image_name: "ghcr.io/openclarity/vmclarity-plugin-kics:latest"
      config: "{}"
```

## SDKs

You can use one of available SDKs in your language of choice to quickly develop scanner plugins for VMClarity.

✅ List of supported languages:
- [Golang](sdk-go)
- [Python](sdk-python)
