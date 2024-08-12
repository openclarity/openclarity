import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { formatDate } from 'utils/utils';

export const getAssetAndScanColumnsConfigList = () => ([
    {
        Header: "Asset name",
        id: "assetName",
        sortIds: ["asset.assetInfo.instanceID"],
        accessor: "asset.assetInfo.instanceID"
    },
    {
        Header: "Asset location",
        id: "assetLocation",
        sortIds: ["asset.assetInfo.location"],
        accessor: "asset.assetInfo.location"
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
