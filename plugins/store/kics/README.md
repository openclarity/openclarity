# KICS

> **KICS** is a scanner application that uses [Checkmarx KICS](https://checkmarx.com/product/opensource/kics-open-source-infrastructure-as-code-project/) (Keeping Infrastructure as Code Secure)
> to scan your Infrastructure as Code (IaC) files for misconfigurations.
> It's designed to be used as a plugin for the [VMClarity](https://openclarity.io/docs/vmclarity/) platform.

## Usage

Make a POST request with config below to the VMClarity API `/assetScans` endpoint to initiate a KICS scan.
The body of the POST request should include a JSON object with the configuration for the scan.

> NOTE: Below is a minimal example. Your actual configuration should have additional properties.

```json
{
    "name": "scan-name",
    "scanTemplate": {
        "scope": "contains(assetInfo.labels, '{\"key\":\"scanconfig\",\"value\":\"test\"}')",
        "assetScanTemplate": {
            "scanFamiliesConfig": {
                "plugins": {
                      // TODO(ramizpolic): Update with request data once decided in plugin integrations work
                }
            }
        }
    }
}
```

### Usage notes

- The KICS scanner is designed to be started by **VMClarity**, therefore running it as a standalone tool is not recommended.

- The value of the `scannerConfig` property in the POST request should contain the [parameters](https://github.com/Checkmarx/kics/blob/e387aa2505a3207e1087520972e0e52f7e0e6fdf/pkg/scan/client.go#L54) that the _KICS_ client will use.

- Please note that not all _scan parameters_ are currently supported by the scanner.

When the scan is done, the output can be found at the `<specified output JSON file>`.
KICS scan findings are exported via `Result` model defined in Scanner Plugin OpenAPI specs.
They are saved to the specified output file in JSON format.

> KICS outputs all its findings as `Misconfiguration` models under `Result.vmclarity.misconfigurations` property.
> See [Scanner Plugin OpenAPI specs](../../openapi.yaml).
