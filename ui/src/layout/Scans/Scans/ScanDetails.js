import React from 'react';
import { useLocation } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import { formatDate, getScanName } from 'utils/utils';
import { ScanDetails as ScanDetailsTab, Findings } from 'layout/detail-displays';
import ScanActionsDisplay from './ScanActionsDisplay';

export const SCAN_DETAILS_PATHS = {
    SCAN_DETALS: "",
    FINDINGS: "findings"
}

const DetailsContent = ({data, fetchData}) => {
    const {pathname} = useLocation();
    
    const {id, name, startTime} = data;

    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Scan details",
                    isIndex: true,
                    component: () => <ScanDetailsTab scanData={data} withAssetScansLink />
                },
                {
                    id: "findings",
                    title: "Findings",
                    path: SCAN_DETAILS_PATHS.FINDINGS,
                    component: () => (
                        <Findings
                            findingsSummary={data?.summary}
                            findingsFilter={`foundBy/scan/id eq '${id}'`}
                            findingsFilterTitle={name}
                        />
                    )
                }
            ]}
            headerCustomDisplay={() => (
                <ScanActionsDisplay data={data} onUpdate={fetchData} />
            )}
            withInnerPadding={false}
        />
    )
}

const ScanDetails = () => (
    <DetailsPageWrapper
        className="scan-details-page-wrapper"
        backTitle="Scans"
        url={APIS.SCANS}
        select="id,name,scanConfig,scope,assetScanTemplate,maxParallelScanners,startTime,endTime,summary,state,stateMessage,stateReason"
        getTitleData={({name, startTime}) => ({title: name, subTitle: formatDate(startTime)})}
        detailsContent={props => <DetailsContent {...props} />}
    />
)

export default ScanDetails;
