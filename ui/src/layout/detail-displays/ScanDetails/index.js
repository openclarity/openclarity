import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import moment from 'moment';
import TitleValueDisplay, { TitleValueDisplayColumn, TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import ScanProgressBar from 'components/ScanProgressBar';
import Button from 'components/Button';
import ConfigurationReadOnlyDisplay from 'layout/Scans/ConfigurationReadOnlyDisplay';
import { formatDate } from 'utils/utils';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import ConfigurationAlertLink from './ConfigurationAlertLink';

export const calculateDuration = (startTime, endTime) => {
    const startMoment = moment(startTime);
    const endMoment = moment(endTime);
    
    const range = ["days", "hours", "minutes", "seconds"].map(item => ({diff: endMoment.diff(startMoment, item), label: item}))
        .find(({diff}) => diff > 1);

    return !!range ? `${range.diff} ${range.label}` : null;
}

const ScanDetails = ({scanData, withAssetScansLink=false}) => {
    const {pathname} = useLocation();
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const {id, scanConfig, scanConfigSnapshot, startTime, endTime, summary, state, stateMessage, stateReason} = scanData || {};
    const {jobsCompleted, jobsLeftToRun} = summary;

    const formattedStartTime = formatDate(startTime);
    
    const onAssetScansClick = () => {
        setFilters(filtersDispatch, {
            type: FILTER_TYPES.ASSET_SCANS,
            filters: {
                filter: `scan/id eq '${id}'`,
                name: `${scanConfigSnapshot.name} - ${formattedStartTime}`,
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
                    <ConfigurationAlertLink updatedConfigData={scanConfig} scanConfigData={scanConfigSnapshot} />
                    <ConfigurationReadOnlyDisplay configData={scanConfigSnapshot} />
                </TitleValueDisplayColumn>
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium>Status</Title>
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
                            <Button onClick={onAssetScansClick}>{`See asset scans (${jobsCompleted || 0})`}</Button>
                        </div>
                    }
                </>
            )}
        />
    )
}

export default ScanDetails;