import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import { AssetDetails, ScanDetails } from 'layout/detail-displays';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabMisconfigurationDetails from './TabMisconfigurationDetails';

const MISCONFIGURATION_DETAILS_PATHS = {
    MISCONFIGURATION_DETAILS: "",
    ASSET_DETAILS: "asset",
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id, asset} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Misconfiguration details",
                    isIndex: true,
                    component: () => <TabMisconfigurationDetails data={data} />
                },
                {
                    id: "asset",
                    title: "Asset details",
                    path: MISCONFIGURATION_DETAILS_PATHS.ASSET_DETAILS,
                    component: () => <AssetDetails assetData={asset} withAssetLink />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const MisconfigurationDetails = () => (
    <FindingsDetailsPage
        backTitle="Misconfigurations"
        getTitleData={({findingInfo}) => ({title: findingInfo.testID})}
        detailsContent={DetailsContent}
    />
)

export default MisconfigurationDetails;
