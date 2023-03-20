import { getScanName } from 'utils/utils';

export const getAssetAndScanColumnsConfigList = () => ([
    {
        Header: "Asset name",
        id: "assetName",
        accessor: "asset.targetInfo.instanceID",
        disableSort: true
    },
    {
        Header: "Asset location",
        id: "assetLocation",
        accessor: "asset.targetInfo.location",
        disableSort: true
    },
    {
        Header: "Scan",
        id: "scan",
        accessor: original => {
            const {scanConfigSnapshot, startTime} = original.scan || {};

            return getScanName({name: scanConfigSnapshot.name, startTime})
        },
        disableSort: true
    }
]);