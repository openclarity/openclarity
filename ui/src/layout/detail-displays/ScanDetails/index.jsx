import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import TitleValueDisplay, { TitleValueDisplayColumn, TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import ScanProgressBar from 'components/ScanProgressBar';
import Button from 'components/Button';
import { SCANS_PATHS } from 'layout/Scans';
import { ScanReadOnlyDisplay } from 'layout/Scans/ConfigurationReadOnlyDisplay';
import { formatDate, calculateDuration, formatNumber } from 'utils/utils';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import ConfigurationAlertLink from './ConfigurationAlertLink';

const ScanDetails = ({scanData, withScanLink=false, withAssetScansLink=false}) => {
    const {pathname} = useLocation();
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const {id, name, scanConfig, scope, assetScanTemplate, startTime, endTime, summary, state, stateMessage, stateReason} = scanData || {};
    const {jobsCompleted, jobsLeftToRun} = summary || {};

    const formattedStartTime = formatDate(startTime);
    
    const onAssetScansClick = () => {
        setFilters(filtersDispatch, {
            type: FILTER_TYPES.ASSET_SCANS,
            filters: {
                filter: `scan/id eq '${id}'`,
                name: name || id,
                suffix: "scan",
                backPath: pathname
            },
            isSystem: true
        });

        navigate(ROUTES.ASSET_SCANS);
    }
    
    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <TitleValueDisplayColumn>
                    <ScanReadOnlyDisplay scanData={scanData} />
                </TitleValueDisplayColumn>
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium onClick={withScanLink ? () => navigate(`${ROUTES.SCANS}/${SCANS_PATHS.SCANS}/${id}`) : undefined}>Scan</Title>
                    <div style={{marginBottom: "20px"}}>
                        <ScanProgressBar
                            state={state}
                            stateReason={stateReason}
                            stateMessage={stateMessage}
                            itemsCompleted={jobsCompleted}
                            itemsLeft={jobsLeftToRun}
                        />
                    </div>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Started">{formattedStartTime}</TitleValueDisplay>
                        <TitleValueDisplay title="Ended">{formatDate(endTime)}</TitleValueDisplay>
                        <TitleValueDisplay title="Duration">{calculateDuration(startTime, endTime)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    {withAssetScansLink &&
                        <div style={{marginTop: "50px"}}>
                            <Title medium>Asset scans</Title>
                            <Button onClick={onAssetScansClick}>{`See asset scans (${formatNumber((jobsCompleted || 0) + (jobsLeftToRun || 0))})`}</Button>
                        </div>
                    }
                </>
            )}
        />
    )
}

export default ScanDetails;
