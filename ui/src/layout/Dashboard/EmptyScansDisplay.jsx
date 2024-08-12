import React from 'react';
import { useNavigate, createSearchParams } from 'react-router-dom';
import EmptyDisplay from 'components/EmptyDisplay';
import { ROUTES } from 'utils/systemConsts';
import { SCANS_PATHS, OPEN_CONFIG_FORM_PARAM } from 'layout/Scans';

const EmptyScansDisplay = () => {
    const navigate = useNavigate();

    const goToScanCongifsPage = params => navigate({
        pathname: `${ROUTES.SCANS}/${SCANS_PATHS.CONFIGURATIONS}`,
        search: createSearchParams(params).toString()
    });

    return (
        <EmptyDisplay
            message={(
                <>
                    <div>No scans detected.</div>
                    <div>Start your first scan to see your VM's issues.</div>
                </>
            )}
            title="New scan configuration"
            onClick={() => goToScanCongifsPage({[OPEN_CONFIG_FORM_PARAM]: true})}
            subTitle="Start scan from config"
            onSubClick={() => goToScanCongifsPage()}
        /> 
    )
}

export default EmptyScansDisplay;