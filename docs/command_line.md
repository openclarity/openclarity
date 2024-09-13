# Initiate scan using the cli

## Reporting results into file

```shell
./bin/openclarity-cli scan --config ~/testConf.yaml -o outputfile
```

If we want to report results to the OpenClarity backend, we need to create asset and asset scan object before scan because it requires asset-scan-id

## Reporting results to OpenClarity backend

```shell
ASSET_ID=$(./bin/openclarity-cli asset-create --file assets/dir-asset.json --server http://localhost:8080/api) --jsonpath {.id}
ASSET_SCAN_ID=$(./bin/openclarity-cli asset-scan-create --asset-id $ASSET_ID --server http://localhost:8080/api) --jsonpath {.id}
./bin/openclarity-cli scan --config ~/testConf.yaml --server http://localhost:8080/api --asset-scan-id $ASSET_SCAN_ID
```

Using one-liner:

```shell
./bin/openclarity-cli asset-create --file docs/assets/dir-asset.json --server http://localhost:8080/api --update-if-exists --jsonpath {.id} | xargs -I{} ./bin/openclarity-cli asset-scan-create --asset-id {} --server http://localhost:8080/api --jsonpath {.id} | xargs -I{} ./bin/openclarity-cli scan --config ~/testConf.yaml --server http://localhost:8080/api --asset-scan-id {}
```
