# Test Plan

## Scan configuration

Create Scan Configuration file

```shell
cat <<EOF > scanconfig.json
{
  "name": "test",
  "scanFamiliesConfig": {
     "sbom": {
       "enabled": true
     },
     "vulnerabilities": {
       "enabled": true
     },
     "exploits": {
       "enabled": true
     }
  },
  "scheduled": {
    "cronLine": "0 */4 * * *",
    "operationTime": "2023-01-20T15:46:18+00:00"
  },
  "scope": {
    "allRegions": true,
    "objectType": "AwsScanScope",
    "instanceTagSelector": [
      {
        "key": "ScanConfig",
        "value": "test"
      }
    ]
  }
}
EOF
```

Apply Scan Configuration to API

```shell
curl -sSf -X POST 'http://localhost:8888/api/scanConfigs' -H 'Content-Type: application/json' \
  -d @scanconfig.json \
| jq -r -e '.id' > scanconfig.id
```

Get Scan Configuration object from API

```shell
curl -sSf -X GET 'http://localhost:8888/api/scanConfigs/'"$(cat scanconfig.id)"'' \
| jq -r -e '.' > scanconfig.api.json
```

## Start Scan

Start Scan using Scan Config

```shell
jq -r -e '{maxParallelScanners, name, scanFamiliesConfig, scheduled, scope} | .scheduled.operationTime = (now|todate)' \
  scanconfig.api.json \
| curl -sSf -X PUT -H 'Content-Type: application/json' 'http://localhost:8888/api/scanConfigs/'"$(cat scanconfig.id)"'' \
  -d @-
```

**Wait until the Scan object is created on the backend/API side**

Get ongoing Scan from API using ScanConfig id

```shell
curl -sSf -G 'http://localhost:8888/api/scans' \
  --data-urlencode "\$filter=scanConfig/id eq '$(cat scanconfig.id)' and state ne 'Done' and state ne 'Failed'" \
| jq -r -e '.items | first' > scan.api.json
```

## Abort Scan in progress

```shell
cat <<EOF > scan-aborted.json
{
  "state": "Aborted"
}
EOF
```

```shell
jq -r -e '.id' scan.api.json > scan.id \
&& curl -sSf -X PATCH -H 'Content-Type: application/json' \
  "http://localhost:8888/api/scans/$(cat scan.id)" \
  -d @scan-aborted.json \
| jq -r -e '.' > scan-aborted.api.json
```
