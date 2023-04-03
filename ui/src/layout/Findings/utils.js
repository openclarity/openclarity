import { formatDate } from 'utils/utils';

export const getAssetAndScanColumnsConfigList = () => ([
    {
        Header: "Asset name",
        id: "assetName",
        sortIds: ["asset.targetInfo.instanceID"],
        accessor: "asset.targetInfo.instanceID"
    },
    {
        Header: "Asset location",
        id: "assetLocation",
        sortIds: ["asset.targetInfo.location"],
        accessor: "asset.targetInfo.location"
    },
    {
        Header: "Scan name",
        id: "scanName",
        sortIds: ["scan.scanConfigSnapshot.name"],
        accessor: "scan.scanConfigSnapshot.name"
    },
    {
        Header: "Scan start",
        id: "startTime",
        sortIds: ["scan.startTime"],
        accessor: original => formatDate(original.scan?.startTime)
    }
]);