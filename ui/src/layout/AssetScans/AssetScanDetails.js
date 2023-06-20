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
    
    const {id, scan, target, summary} = data;
    const {id: scanId, scanConfigSnapshot, startTime} = scan;
    const {id: targetId, targetInfo} = target;
    
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
                            findingsFilter={`scan/id eq '${scanId}' and asset/id eq '${targetId}'`}
                            findingsFilterTitle={`${targetInfo.instanceID} scanned by ${getScanName({name: scanConfigSnapshot.name, startTime})}`}
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
        select="id,scan,target,summary,status"
        expand="scan($select=id,scanConfigSnapshot,startTime,endTime),target($select=id,targetInfo),status"
        getTitleData={({scan, target}) => {
            const {startTime, scanConfigSnapshot} = scan || {};

            return ({
                title: target?.targetInfo?.instanceID,
                subTitle: `scanned by '${scanConfigSnapshot?.name}' on ${formatDate(startTime)}`
            })
        }}
        detailsContent={props => <DetailsContent {...props} />}
        withPadding
    />
)

export default AssetScanDetails;