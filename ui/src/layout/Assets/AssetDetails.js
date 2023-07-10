import React from 'react';
import { useLocation } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import { AssetDetails as AssetDetailsTab, Findings } from 'layout/detail-displays';

const ASSET_DETAILS_PATHS = {
    ASSET_DETAILS: "",
    FINDINGS: "findings"
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id, assetInfo, summary} = data || {};
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Asset details",
                    isIndex: true,
                    component: () => <AssetDetailsTab assetData={data} withAssetScansLink />
                },
                {
                    id: "findings",
                    title: "Findings",
                    path: ASSET_DETAILS_PATHS.FINDINGS,
                    component: () => (
                        <Findings
                            findingsSummary={summary}
                            findingsFilter={`asset/id eq '${id}'`}
                            findingsFilterTitle={assetInfo.instanceID}
                            findingsFilterSuffix="asset"
                        />
                    )
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const AssetDetails = () => (
    <DetailsPageWrapper
        backTitle="Assets"
        url={APIS.ASSETS}
        select="id,assetInfo,summary,firstSeen,lastSeen,terminatedOn"
        getTitleData={({assetInfo}) => ({title: assetInfo?.instanceID})}
        detailsContent={props => <DetailsContent {...props} />}
        withPadding
    />
)

export default AssetDetails;
