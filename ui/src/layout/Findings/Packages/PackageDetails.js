import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import { AssetDetails, ScanDetails } from 'layout/detail-displays';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabPackageDetails from './TabPackageDetails';

const PACKAGE_DETAILS_PATHS = {
    PACKAGE_DETAILS: "",
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
                    title: "Package details",
                    isIndex: true,
                    component: () => <TabPackageDetails data={data} />
                },
                {
                    id: "asset",
                    title: "Asset details",
                    path: PACKAGE_DETAILS_PATHS.ASSET_DETAILS,
                    component: () => <AssetDetails assetData={asset} withAssetLink />
                },
                {
                    id: "scan",
                    title: "Scan details",
                    path: PACKAGE_DETAILS_PATHS.SCAN_DETAILS,
                    component: () => <ScanDetails scanData={scan} withScanLink />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const PackageDetails = () => (
    <FindingsDetailsPage
        backTitle="Packages"
        getTitleData={({findingInfo}) => ({title: findingInfo.name, subTitle: findingInfo.version})}
        detailsContent={DetailsContent}
    />
)

export default PackageDetails;