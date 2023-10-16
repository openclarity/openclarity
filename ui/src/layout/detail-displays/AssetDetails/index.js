import React, { useMemo } from 'react';
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
import { getAssetName } from 'layout/Assets/utils';

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

    const {id, assetInfo, firstSeen, lastSeen, terminatedOn} = assetData;
    const {objectType, location, tags, labels, image, instanceType, platform, launchTime, rootVolume, repoDigests, createdAt} = assetInfo || {};
    const {sizeGB, encrypted} = rootVolume || {};

    const platformFormatted = useMemo(() => {
        const getImagePlatform = (image) => image.os && image.architecture ? `${image.os}/${image.architecture}` : null;
        return platform ?? getImagePlatform(assetInfo) ?? getImagePlatform(image);
    }, [assetInfo, image, platform]);
    
    const imageFormatted = useMemo(() => {
        if (typeof image === "string") {
            return image;
        }

        const getImageName = (image) => image.repoTags?.[0] || image.imageID;

        return typeof image === "object" ? getImageName(image) : getImageName(assetInfo);
    }, [assetInfo, image]);

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <Title medium onClick={withAssetLink ? () => navigate(`${ROUTES.ASSETS}/${id}`) : undefined}>Asset</Title>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Name">{getAssetName(assetInfo)}</TitleValueDisplay>
                        <TitleValueDisplay title="Type">{objectType}</TitleValueDisplay>
                        <TitleValueDisplay title="Location">{location || repoDigests?.[0]}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="First Seen">{formatDate(firstSeen)}</TitleValueDisplay>
                        <TitleValueDisplay title="Last Seen">{formatDate(lastSeen)}</TitleValueDisplay>
                        <TitleValueDisplay title="Terminated On">{formatDate(terminatedOn)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Labels"><TagsList items={formatTagsToStringsList(tags ?? labels)} /></TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Image">{imageFormatted}</TitleValueDisplay>
                        <TitleValueDisplay title="Instance type">{instanceType}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Platform">{platformFormatted}</TitleValueDisplay>
                        <TitleValueDisplay title="Launch time">{formatDate(launchTime || createdAt)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Root Volume Size">{sizeGB} GB</TitleValueDisplay>
                        <TitleValueDisplay title="Encrypted Root Volume">{encrypted}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>
            )}
            rightPlaneDisplay={!withAssetScansLink ? null : () => <AssetScansDisplay assetName={getAssetName(assetInfo)} assetId={id} />}
        />
    )
}

export default AssetDetails;
