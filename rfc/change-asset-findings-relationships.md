# [RFC] Extend asset-finding relationship logic

*Note: this RFC template follows HashiCrop RFC format described [here](https://works.hashicorp.com/articles/rfc-template)*


|               |                                                               |
|---------------|---------------------------------------------------------------|
| **Created**   | 2023-01-15                                                    |
| **Status**    | WIP\| InReview \| **Approved** \| Obsolete                    |
| **Owner**     | *ramizpolic*                                                  |
| **Approvers** | [PR-1042](https://github.com/openclarity/vmclarity/pull/1042) |

---

This RFC proposes adding `AssetFinding` API model and its supporting logic to achieve API consistency, improve efficiency, and enable aggregation.

## Background

Each finding defines some specific security details (e.g. vulnerability) discovered on an asset.
The same finding can be discovered on multiple assets by different asset scans.
This means that there is many-to-many relationship between findings, assets, and asset scans.

However, the existing [specifications](https://github.com/openclarity/vmclarity/blob/9aa03a8abe22ebddb841a9c28f7a9629f744ced7/api/openapi.yaml#L3395-L3444)
describe findings with one-to-one relationship to assets and asset scans.
In addition, the [database logic](https://github.com/openclarity/vmclarity/blob/9aa03a8abe22ebddb841a9c28f7a9629f744ced7/pkg/apiserver/database/gorm/finding.go#L103-L105)
treats every new finding as unique without performing any checks.
Together, this can introduce issues for multiple reasons:

- Each `Finding` is coupled with an `Asset` it was discovered on and the `AssetScan` that discovered it.
  This leads to unnecessary data duplication as each `Finding` with different associations to these models will be treated as unique.
  In addition, the lack of an actual uniqueness check completely ignores the data already present in the database.
  Together, this creates performance and memory utilization overheads.
- The model differs from the existing association patterns between models compared to e.g. `Asset`, `AssetScan`, and `AssetScanEstimation`.
  This introduces complexities due to a lack of proper aggregation by overloading the `Finding` data returned by the API or shown on the UI.

## Proposal

Findings should serve as a collection of all security details discovered so far, irrelevant of the assets they were discovered on or the scans that discovered them.
This means that the _existing `Asset` relationship can be dropped from the `Finding` model_.
Instead, to express the relationship between assets and findings, _new `AssetFinding` model should be added_.
This ensures many-to-many relationship between the two, i.e. a single finding can be discovered on multiple assets, and an asset can contain multiple findings.

A similar approach can be used for asset scans and findings, although this is less relevant (check [non-goals](#non-goals) section).
Instead, _the `AssetScan` relationship in `Finding` should be preserved, but it should be noted that it represents the first scan that discovered a given finding._

The database logic should _implement uniqueness check_ similar to the existing logic as shown [here](https://github.com/openclarity/vmclarity/blob/9aa03a8abe22ebddb841a9c28f7a9629f744ced7/pkg/apiserver/database/gorm/asset.go#L289).
The data required for the check can be extracted directly from the actual object.
Uniqueness checks are required for both `Finding` and `AssetFinding` models.

To provide statistical data and aggregation capabilities between other models, _`Finding` should be expanded with the `assetsCount` property_.
Together, these approaches address the API consistency and data duplication issues.

### Analysis

Take the number of packages that can be discovered on a single production container image as a reference.
In a real-world scenario, many security findings can be discovered on a specific asset.
Additionally, assets and findings are versioned models where every new version is treated as a completely new object.
Similarly, asset scans are also versioned in terms of schedule, i.e. new asset scans can be created weekly.

Due to the current nature of these models (e.g. creating a new `Finding` or `Asset` for each version, or `AssetScan` on schedule), the size of the tables can grow rapidly.
The reason to have `AssetFinding` and `AssetScanFinding` as separate tables relates to time and space complexity.
This is also why a unified table between all three models is not viable.

**Example A**
To keep track of `#assets = 100`, `#assetScans = 100`, and `#findings = 100` with no versioning changes and no additional scans requires:
```
R(asset, assetScan, finding) = #asset * #assetScan * #finding = 100^3 items - Unified table is larger than both other tables combined
R(asset, finding)            = #asset * #finding              = 100^2 items - AssetFinding table
R(assetScan, finding)        = #assetScan * #finding          = 100^2 items - AssetScanFinding table
```

**Example B**
Unlike the previous example, this takes into account the versioning and scheduling which is a better metric for real-world usage.
To keep track of `#assets = 100`, `#assetScans = 100`, and `#findings = 100` with _10 new versions_ and _10 new assets for each scan_ requires:
```
R(asset, assetScan, finding, version, schedule) = 10^5 * 100^2 - Combined relationship table grows too quickly compared to other tables
R(asset, finding, version)                      = 10^2 * 100^2 - AssetFinding table depends on assets and findings, i.e. versioning; also transiently depends on scans, but this is omitted for simplicity.
R(assetScan, finding, version, schedule)        = 10^2 * 100^2 - Scans can be scheduled to occur more often than the versioning changes on other models
```

Therefore, it makes sense to keep track of associations in separate tables between these models to reduce size and improve efficiency.
Otherwise, sharding or more complex approaches must be used which are not viable options at this stage.

### Non-goals

This RFC does not intend to propose changes regarding the relationship between findings and asset scans.
It is assumed that the asset scan in a given finding represents the last scan that discovered it, while the timestamps are used for object lifecycle tracking.
If required to keep track of all asset scans that discovered a specific finding and vice-versa, a new `AssetScanFinding` model can be added.

Apart from statistical analysis, this offers no additional benefit as we usually do not care much about the actual scans but rather their results in terms of discovery of new assets and findings.
This behavior, if required, can be addressed following the same approach as described in the [Addressing API changes](#addressing-api-changes) section.

### Abandoned Ideas (Optional)

Adding the aggregation methods to the `uibackend` API was considered but abandoned as it does not address the data duplication issue.

---

## Implementation

### 1. Extend Findings API model and the supporting logic

The new models should be described as follows:

```yaml
Finding:
  type: object
  allOf:
    - $ref: '#/components/schemas/Metadata'
    - type: object
      properties:
        id:
          type: string
        revision: # added in case some of the non-unique identifiers are updated
          type: integer
        firstSeen:
          description: When this finding was first discovered by a scan
          type: string
          format: date-time
        lastSeen:
          description: When this finding was last discovered by a scan
          type: string
          format: date-time
        lastSeenBy:
          # represents the last scan that discovered this finding
          # TODO(ramizpolic): maybe ScanRelationship could be used instead to avoid unnecessary object updates
          $ref: '#/components/schemas/AssetScanRelationship'
        assetsCount:
          description: Total number of assets that contain this finding
          type: integer
        # invalidatedOn: # removes property and uses `assetsCount > 0` check for validity
        #  description: When this finding was invalidated by a newer scan
        #  type: string
        #  format: date-time
        findingInfo:
          anyOf:
            - $ref: '#/components/schemas/PackageFindingInfo'
            - $ref: '#/components/schemas/VulnerabilityFindingInfo'
            - $ref: '#/components/schemas/MalwareFindingInfo'
            - $ref: '#/components/schemas/SecretFindingInfo'
            - $ref: '#/components/schemas/MisconfigurationFindingInfo'
            - $ref: '#/components/schemas/RootkitFindingInfo'
            - $ref: '#/components/schemas/ExploitFindingInfo'
            - $ref: '#/components/schemas/InfoFinderFindingInfo'
          discriminator:
            propertyName: objectType
            mapping:
              Package: '#/components/schemas/PackageFindingInfo'
              Vulnerability: '#/components/schemas/VulnerabilityFindingInfo'
              Malware: '#/components/schemas/MalwareFindingInfo'
              Secret: '#/components/schemas/SecretFindingInfo'
              Misconfiguration: '#/components/schemas/MisconfigurationFindingInfo'
              Rootkit: '#/components/schemas/RootkitFindingInfo'
              Exploit: '#/components/schemas/ExploitFindingInfo'
              InfoFinder: '#/components/schemas/InfoFinderFindingInfo'

## FindingRelationship should be added similarly to other read-only models
# FindingRelationship:
# ...
```

The API changes impact the database schema and should be handled accordingly.
In addition, the database-related logic such as bootstrapping the demo data needs to be updated to reflect these changes.
For the current case, assume that the `assetsCount` is always zero.

### 2. Add uniqueness checks to Findings database model

The uniqueness check can be added similarly to the existing implementations. An example is given below.

```go
func (s *FindingsTableHandler) checkUniqueness(finding models.Finding) (*models.Finding, error) {
  discriminator, err := finding.FindingInfo.ValueByDiscriminator()
  if err != nil {
    return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
  }
  
  var filter string
  switch info := discriminator.(type) {
  case models.PackageFindingInfo:
    filter = fmt.Sprintf("uniqueness query for package finding", info)
  
  case models.VulnerabilityFindingInfo:
    filter = fmt.Sprintf("uniqueness query for vulnerability finding", info)
  
  // implementation of other cases
  
  default:
    return nil, fmt.Errorf("finding type is not supported (%T): %w", discriminator, err)
  }
  
  // implementation of the actual check
  
  return nil, nil
}
```

### 3. Add AssetFindings API model and the supporting logic

```yaml
AssetFinding:
  type: object
  allOf:
    - $ref: '#/components/schemas/Metadata'
    - type: object
      properties:
        id:
          type: string
        asset:
          $ref: '#/components/schemas/AssetRelationship'
        finding:
          $ref: '#/components/schemas/FindingRelationship'
        # `ignored` property can be added to allow users to individually ignore specific security findings on a given asset.
        # This is not relevant for this RFC as it only serves as an extension example.
```

The implementation should address the following cases:
- Update database schema and implement related database table. Additional table index should be created for `(asset.ID, finding.ID)` fields.
- Add uniqueness check for the new database model
- Add CRUD controller logic for the new model to handle `/asset-findings` and `/asset-finding/{id}` routes.
  - The filters for `GET /asset-finding` can contain either `findingID` or `assetID` to get the list of related objects for a given filter.
    This allows fetching the objects for a query such as _which assets/findings are associated for given finding/asset_.
    The UI can make use of this data by extracting only the required fields (e.g. `.asset`) from the `[]AssetFinding` response.
  - When performing CRUD operations, note that the `assetsCount` property of the related `Finding` can change either by decrementing or incrementing it.

### 4. Update related UI components

Described in [UI](#ui) section.

## UX

This RFC has no visible impacts on the UX.

## UI

This RFC changes the following UI components:
- `/findings/{findingType}` table drops the _Asset Name_ and _Asset location_ columns.
  Instead, they are replaced with a single _Asset Count_ column that shows the number of assets related to a given finding.
  The `firstSeen` field needs to be used to describe when the asset was discovered. 
- `/findings/{findingType}/{findingID}` changes _Asset details_ menu item by showing the table instead of a single asset.
  The relevant data is obtained by fetching `[]AssetFinding` from `GET /asset-findings` for a given `findingID`.
  Subsequently, the asset data for the table is extracted from the returned array by extracting the `.asset` key.
