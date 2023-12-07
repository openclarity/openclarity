import React from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import { APIS } from 'utils/systemConsts';
import ConfigurationActionsDisplay from '../ConfigurationActionsDisplay';
import TabConfiguration from './TabConfiguration';

import './configuration-details.scss';

export const SCAN_CONFIGS_SCAN_TAB_PATH = "scans";

const DetailsContent = ({data, fetchData}) => {
    const navigate = useNavigate();
    const {pathname} = useLocation();
    const params = useParams();

    const id = params["id"];
    const innerTab = params["*"];

    return (
        <div className="configuration-details-content">
            <div className="configuration-details-content-header">
                <ConfigurationActionsDisplay
                    data={data}
                    onDelete={() => navigate(pathname.replace(`/${id}/${innerTab}`, ""))}
                    onUpdate={fetchData}
                />
            </div>
            <TabConfiguration data={data} />
        </div>
    )
}

const ConfigurationDetails = () => (
    <DetailsPageWrapper
        className="configuration-details-page-wrapper"
        backTitle="Scan configurations"
        url={APIS.SCAN_CONFIGS}
        select="id,name,scanTemplate/scope,scanTemplate/assetScanTemplate/scanFamiliesConfig,scheduled,scanTemplate/maxParallelScanners,scanTemplate/assetScanTemplate/scannerInstanceCreationConfig"
        getTitleData={data => ({title: data?.name})}
        detailsContent={DetailsContent}
    />
)

export default ConfigurationDetails;
