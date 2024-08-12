import React from 'react';
import { useNavigate } from 'react-router-dom';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import { ROUTES } from 'utils/systemConsts';
import { getAssetName } from 'utils/utils';
import { VMInfoDetails } from './VMInfoDetails';
import { ContainerImageInfoDetails } from './ContainerImageInfoDetails';
import { ContainerInfoDetails } from './ContainerInfoDetails';
import { AssetScansDisplay } from './AssetScansDisplay';
import { PodInfoDetails } from './PodInfoDetails';
import { DirInfoDetails } from './DirInfoDetails';
import { CommonAssetMetadata } from './CommonAssetMetadata';

const AssetDetailsByType = ({ assetData }) => {
    if (!assetData?.assetInfo) {
        return null;
    }

    switch (assetData.assetInfo.objectType) {
        case 'VMInfo':
            return <VMInfoDetails assetData={assetData} />;
        case 'ContainerImageInfo':
            return <ContainerImageInfoDetails assetData={assetData} />;
        case 'ContainerInfo':
            return <ContainerInfoDetails assetData={assetData} />;
        case 'PodInfo':
            return <PodInfoDetails assetData={assetData} />;
        case 'DirInfo':
            return <DirInfoDetails assetData={assetData} />;
        default:
            return null;
    }
}

const AssetDetails = ({assetData, withAssetLink=false, withAssetScansLink=false}) => {
    const navigate = useNavigate();

    const {id, assetInfo} = assetData;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <Title medium onClick={withAssetLink ? () => navigate(`${ROUTES.ASSETS}/${id}`) : undefined}>Asset</Title>
                    <CommonAssetMetadata assetData={assetData} />
                    <AssetDetailsByType assetData={assetData} />
                </>
            )}
            rightPlaneDisplay={!withAssetScansLink ? null : () => <AssetScansDisplay assetName={getAssetName(assetInfo)} assetId={id} />}
        />
    )
}

export default AssetDetails;
