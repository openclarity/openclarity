import React from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import TabbedPage from 'components/TabbedPage';
import { APIS } from 'utils/systemConsts';
import ConfigurationActionsDisplay from '../ConfigurationActionsDisplay';
import TabConfiguration from './TabConfiguration';
import TabScans from './TabScans';

import './configuration-details.scss';

export const SCAN_CONFIGS_SCAN_TAB_PATH = "scans";

const getReplace = params => {
    const id = params["id"];
    const innerTab = params["*"];
    
    return !!innerTab ? `/${id}/${innerTab}` : `/${id}`;
}

const DetailsContent = ({data}) => {
    const navigate = useNavigate();
    const {pathname} = useLocation();
    const params = useParams();
    
    const {id} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "config",
                    title: "Configuration",
                    isIndex: true,
                    component: () => <TabConfiguration data={data} />
                },
                {
                    id: "scans",
                    title: "Scans",
                    path: SCAN_CONFIGS_SCAN_TAB_PATH,
                    component: () => <TabScans data={data} />
                }
            ]}
            headerCustomDisplay={() => (
                <ConfigurationActionsDisplay
                    data={data}
                    onDelete={() => navigate(pathname.replace(getReplace(params), ""))}
                />
            )}
            withInnerPadding={false}
        />
    )
}

const ConfigurationDetails = () => (
    <DetailsPageWrapper
        className="configuration-details-page-wrapper"
        backTitle="Scan configurations"
        url={APIS.SCAN_CONFIGS}
        getTitleData={data => ({title: data?.name})}
        detailsContent={DetailsContent}
        getReplace={params => getReplace(params)}
    />
)

export default ConfigurationDetails;