import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { formatDate, getAssetName } from 'utils/utils';

const NAME_SORT_IDS = [
    "asset.assetInfo.instanceID",
    "asset.assetInfo.podName",
    "asset.assetInfo.dirName",
    "asset.assetInfo.imageID",
    "asset.assetInfo.containerName"
];

export const getAssetAndScanColumnsConfigList = () => ([
    {
        Header: "Asset name",
        id: "assetName",
        sortIds: NAME_SORT_IDS,
        accessor: (data) => getAssetName(data.asset.assetInfo),
    },
    {
        Header: "Asset location",
        id: "assetLocation",
        sortIds: ["asset.assetInfo.location"],
        accessor: (data) => data.asset.assetInfo.location || data.asset.assetInfo.repoDigests?.[0],
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
                <TitleValueDisplay title="Invalidated on">{formatDate(invalidatedOn)}</TitleValueDisplay>
            </TitleValueDisplayRow>
        }
    </>
)
