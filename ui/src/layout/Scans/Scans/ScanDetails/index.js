import React from 'react';
import { useLocation } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import { formatDate } from 'utils/utils';
// import ScanActionsDisplay from '../ScanActionsDisplay';
import TabGeneral from './TabGeneral';
import TabFindings from './TabFindings';

import './scan-details.scss';

export const SCANS_FINDINGS_TAB_PATH = "findings";

const getReplace = params => {
    const id = params["id"];
    const innerTab = params["*"];
    
    return !!innerTab ? `/${id}/${innerTab}` : `/${id}`;
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "General",
                    isIndex: true,
                    component: () => <TabGeneral data={data} />
                },
                {
                    id: "findings",
                    title: "Findings",
                    path: SCANS_FINDINGS_TAB_PATH,
                    component: () => <TabFindings data={data} />
                }
            ]}
            // headerCustomDisplay={() => (
            //     <ScanActionsDisplay data={data} />
            // )}
            withInnerPadding={false}
        />
    )
}

const ScanDetails = () => (
    <DetailsPageWrapper
        className="scan-details-page-wrapper"
        backTitle="Scans"
        getUrl={({id}) => `${APIS.SCANS}/${id}?$expand=ScanConfig`}
        getTitleData={({scanConfigSnapshot, startTime}) => ({title: scanConfigSnapshot?.name, subTitle: formatDate(startTime)})}
        detailsContent={props => <DetailsContent {...props} />}
        getReplace={params => getReplace(params)}
    />
)

export default ScanDetails;