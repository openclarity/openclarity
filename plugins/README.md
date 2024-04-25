# Overview

**Plugins** provide additional **scanning capabilities** to VMClarity ecosystem.
_Plugins are executed as Docker containers._
Project structure is defined as:

- **runner** - Provides necessary logic to execute scanner plugins. Used as a library in VMClarity to enable plugin features.
- **sdk** - Language-specific libraries, templates, and examples to help with the implementation of scanner plugins.
Import these packages in custom implementation of scanner plugins for a quick start.
- **store** - Collection of implemented containerized scanner plugins.
Plugins are publicly available as container images that can be with VMClarity.

### Support

Plugin feature is supported on machines that run Docker Daemon on the same host as VMClarity CLI.
Note that the user who runs the CLI needs to have access to the Docker Daemon in order for plugins feature to work.

✅ List of supported environments:
1. AWS
2. GCP
3. Azure
4. Docker

❌ List of unsupported environments:
- _Kubernetes_ - We plan on adding plugin support for Kubernetes. Background: Security is important and tricky to handle when exposing daemon to the CLI container.

### Usage

You can start using plugins via **[Plugins Store](store)**.
For example, you can pass the `.families.yaml` config file defined below to the CLI scan command.
This configuration uses **KICS scanner** to scan `/tmp` dir for security misconfigurations.

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
      config: "{\"customKey\":\"customValue\"}"
```

### SDKs
You can quickly develop scanner plugins that can be used in VMClarity using available SDKs.

✅ List of supported languages:
- [Golang](sdk/go)
- [Python](sdk/python)
