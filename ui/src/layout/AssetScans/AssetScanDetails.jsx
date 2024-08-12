import React from 'react';
import { useLocation } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import { formatDate, getScanName } from 'utils/utils';
import { Findings } from 'layout/detail-displays';
import TabAssetScanDetails from './TabAssetScanDetails';

const ASSET_SCAN_DETAILS_PATHS = {
    ASSET_SCAN_DETAILS: "",
    FINDINGHS: "findings"
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id, summary} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Asset scan details",
                    isIndex: true,
                    component: () => <TabAssetScanDetails data={data} />
                },
                {
                    id: "findings",
                    title: "Findings",
                    path: ASSET_SCAN_DETAILS_PATHS.FINDINGHS,
                    component: () => (
                        <Findings
                            findingsSummary={summary}
                            findingsFilter={`foundBy/id eq '${id}'`}
                            findingsFilterTitle={`AssetScan ${id}`} // TODO(sambetts) replace with name maybe
                        />
                    )
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const AssetScanDetails = () => (
    <DetailsPageWrapper
        backTitle="Asset scans"
        url={APIS.ASSET_SCANS}
        select="id,scan,asset,summary,status,stats,sbom,vulnerabilities,exploits,misconfigurations,secrets,malware,rootkits"
        expand="scan($select=id,name,startTime,endTime),asset($select=id,assetInfo),status,stats"
        getTitleData={({scan, asset}) => {
            const {startTime, name} = scan || {};

            return ({
                title: asset?.assetInfo?.instanceID,
                subTitle: `scanned by '${name}' on ${formatDate(startTime)}`
            })
        }}
        detailsContent={props => <DetailsContent {...props} />}
        withPadding
    />
)

export default AssetScanDetails;
