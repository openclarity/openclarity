import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useFetch } from 'hooks';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import Button from 'components/Button';
import Loader from 'components/Loader';
import { TagsList } from 'components/Tag';
import { ROUTES, APIS } from 'utils/systemConsts';
import { formatNumber, formatDate, formatTagsToStringsList } from 'utils/utils';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';

const AssetScansDisplay = ({assetName, assetId}) => {
    const {pathname} = useLocation();
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const filter = `asset/id eq '${assetId}'`;
    
    const onAssetScansClick = () => {
        setFilters(filtersDispatch, {
            type: FILTER_TYPES.ASSET_SCANS,
            filters: {filter, name: assetName, suffix: "asset", backPath: pathname},
            isSystem: true
        });

        navigate(ROUTES.ASSET_SCANS);
    }
    
    const [{loading, data, error}] = useFetch(APIS.ASSET_SCANS, {
        queryParams: {"$filter": filter, "$count": true, "$select": "id,asset,summary,scan"}
    });
    
    if (error) {
        return null;
    }

    if (loading) {
        return <Loader absolute={false} small />
    }
    
    return (
        <>
            <Title medium>Asset scans</Title>
            <Button onClick={onAssetScansClick} >{`See all asset scans (${formatNumber(data?.count || 0)})`}</Button>
        </>
    )
}

const AssetDetails = ({assetData, withAssetLink=false, withAssetScansLink=false}) => {
    const navigate = useNavigate();

    const {id, assetInfo, lastSeen, terminated} = assetData;
    const {instanceID, objectType, location, tags, image, instanceType, platform, launchTime, rootVolumeSizeGB, rootVolumeEncrypted} = assetInfo || {};
    
    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <Title medium onClick={withAssetLink ? () => navigate(`${ROUTES.ASSETS}/${id}`) : undefined}>Asset</Title>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Name">{instanceID}</TitleValueDisplay>
                        <TitleValueDisplay title="Type">{objectType}</TitleValueDisplay>
                        <TitleValueDisplay title="Last Seen">{formatDate(lastSeen)}</TitleValueDisplay>
                        <TitleValueDisplay title="Terminated">{formatDate(terminated)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Location">{location}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Labels"><TagsList items={formatTagsToStringsList(tags)} /></TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Image">{image}</TitleValueDisplay>
                        <TitleValueDisplay title="Instance type">{instanceType}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Platform">{platform}</TitleValueDisplay>
                        <TitleValueDisplay title="Launch time">{formatDate(launchTime)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Root Volume Size">{rootVolumeSizeGB} GB</TitleValueDisplay>
                        <TitleValueDisplay title="Encrypted Root Volume">{rootVolumeEncrypted ? "Yes" : "No"}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>
            )}
            rightPlaneDisplay={!withAssetScansLink ? null : () => <AssetScansDisplay assetName={instanceID} assetId={id} />}
        />
    )
}

export default AssetDetails;
