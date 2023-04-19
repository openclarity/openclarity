import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
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
    },
    {
        Header: "Found on",
        id: "foundOn",
        sortIds: ["foundOn"],
        accessor: original => formatDate(original.foundOn)
    }
]);

export const FindingsDetailsCommonFields = ({foundOn, invalidatedOn}) => (
    <>
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Found on">{formatDate(foundOn)}</TitleValueDisplay>
        </TitleValueDisplayRow>
        {!!invalidatedOn &&
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Replaced by a newer scan on">{formatDate(invalidatedOn)}</TitleValueDisplay>
            </TitleValueDisplayRow>
        }
    </>
)