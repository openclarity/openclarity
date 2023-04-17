import React from 'react';
import { useLocation } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import { AssetDetails, ScanDetails } from 'layout/detail-displays';
import TabMisconfigurationDetails from './TabMisconfigurationDetails';

const MISCONFIGURATION_DETAILS_PATHS = {
    MISCONFIGURATION_DETAILS: "",
    ASSET_DETAILS: "asset",
    SCAN_DETAILS: "scan"
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id, scan, asset} = data;
    
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
                    component: () => <AssetDetails assetData={asset} />
                },
                {
                    id: "scan",
                    title: "Scan details",
                    path: MISCONFIGURATION_DETAILS_PATHS.SCAN_DETAILS,
                    component: () => <ScanDetails scanData={scan} />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const MisconfigurationDetails = () => (
    <DetailsPageWrapper
        backTitle="Misconfigurations"
        getUrl={({id}) => `${APIS.FINDINGS}/${id}?$expand=asset,scan`}
        getTitleData={({findingInfo}) => ({title: findingInfo.testID})}
        detailsContent={props => <DetailsContent {...props} />}
    />
)

export default MisconfigurationDetails;