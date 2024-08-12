# [RFC] Extend misconfiguration scanners

*Note: this RFC template follows HashiCrop RFC format described [here](https://works.hashicorp.com/articles/rfc-template)*


|               |                                                               |
|---------------|---------------------------------------------------------------|
| **Created**   | 2024-01-19                                                    |
| **Status**    | WIP\| InReview \| **Approved** \| Obsolete                    |
| **Owner**     | *ramizpolic*                                                  |
| **Approvers** | [PR-1114](https://github.com/openclarity/vmclarity/pull/1114) |

---

This RFC proposes the extension of misconfiguration scanning logic to integrate CIS Docker Benchmark and enrich security findings on assets.

## Background

> [!NOTE]
> The scanning logic relies on using explicit scopes such as vulnerabilities and misconfigurations to categorize security findings on assets.
Generally, this works well when scanners have a well-defined boundary used to determine a specific scope.
However, some scanners cannot directly categorize findings using a single or existing scope which can limit integration options.
This behavior, along with the lack of dynamic- and multi-scope options, also underlines an important limitation of how findings are being described, categorized, processed, and analyzed.
Note that this RFC does not intend to resolve this behavior, but rather draw attention to it.

The integration of [CIS Docker Benchmark](https://github.com/goodwithtech/dockle) scanner requires additional changes to address the scope-based categorization limitations.
In KubeClarity, the scanner defines its own [API model](https://github.com/openclarity/kubeclarity/blob/5ac3048b7a782c900a9bef846a91a7735ba77e24/api/swagger.yaml#L243C26-L243C26) to describe related security findings.
This makes the migration of scanning logic to VMClarity problematic for two main reasons:

- Logic in the form of a new independent scanner family does not conform to any existing *security scopes*.
CIS Docker Benchmark provides little benefit on its own due to scope constraints currently defined for the scanning logic.

- Logic is *too specific* and *provider-dependant* to be part of an existing scanner family.
CIS Docker Benchmark scan results cannot be uniformly converted to other findings without some loss of data.

## Proposal

The CIS Docker Benchmark scanner can be migrated as part of **misconfiguration scanner family** to enrich the findings on assets with additional security coverage.
Contextually, the misconfiguration findings serve as a superset of CIS Docker Benchmark results.
This approach benefits VMClarity in several ways:

* The misconfiguration findings can be generalized and reused without impacting the existing scopes

The misconfiguration [API model](https://github.com/openclarity/vmclarity/blob/bfc32ec88ee266157aaf7bcae7b17c4b2ee5c868/api/openapi.yaml#L3083) is not abstract enough to enable integration of new scanners.
Minor API changes are required to make the model more generic and enable direct conversion of CIS Docker Benchmark results.
This also standardizes the model for usage and simplifies future integrations.

- The misconfiguration scanner family enables an idiomatic way to migrate the required scanning logic from KubeClarity

Integrating the CIS Docker Benchmark can be accomplished by reusing the existing patterns to minimize changes.
The migration can then be performed as an implementation of a scanner within the existing misconfiguration family.

### Abandoned Ideas (Optional)

---

## Implementation

1. Extend the misconfiguration model to be more generic

The model preserves most of its original properties but is generalized to enable direct conversion from CIS Docker Benchmark results.

```yaml
Misconfiguration:
  type: object
  properties:
    scannerName: # preserved
      type: string
    id: # replaces `testID`; maps CISDockerBenchmarkResultsEX.code
      type: string
      description: Check or test ID, if applicable (e.g. Lynis TestID, CIS Docker Benchmark checkpoint code, etc)
    location: # replaces `scannedPath`; maps from the underlying data returned by the CIS Docker Benchmark scanner
      type: string
      description: Location within the asset where the misconfiguration was recorded (e.g. filesystem path)
    category: # replaces `testCategory`; uses static `best-practice` to label CIS Docker Benchmark results.
              # Additional categories such as `security` can be extracted/mapped, but not relevant to this RFC.
      type: string
      description: Specifies misconfiguration impact category
    message: # preserved; maps CISDockerBenchmarkResultsEX.title
      type: string
      description: Short info about the misconfiguration
    description: # replaces `testDescription`; maps CISDockerBenchmarkResultsEX.desc
      type: string
      description: Additional context such as the potential impact
    remediation: # preserved
      type: string
      description: Possible fix for the misconfiguration
    severity: # preserved; maps CISDockerBenchmarkResultsEX.level
      $ref: '#/components/schemas/MisconfigurationSeverity'
```

2. Update related UI components to support the API changes

3. Migrate CIS Docker Benchmark scanner from KubeClarity as part of misconfiguration scanners

The [scanner code](https://github.com/openclarity/kubeclarity/tree/5f6b411161100c15196c8149c0b1df5537c88a05/cis_docker_benchmark_scanner) defined in KubeClarity can be migrated under the misconfiguration scanner family following the existing patterns.
Note that minor changes are required to ensure that the results returned by the scanner conform to the new misconfiguration API model.
_Lynis_ misconfiguration scanner can serve as a reference on required code changes for successful integration. 

## UX

This RFC has no visible impacts on the UX.

## UI

This RFC changes the `Finding` components related to misconfigurations shown on the UI by using updated API models.
