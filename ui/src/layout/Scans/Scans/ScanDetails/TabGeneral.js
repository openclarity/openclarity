import React from 'react';
import TitleValueDisplay, { TitleValueDisplayColumn, TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import { ScopeDisplay, ScanTypesDisplay, InstancesDisplay } from 'layout/Scans/scopeDisplayUtils';
import { formatDate, calculateDuration } from 'utils/utils';
import ScanStatusDisplay from '../ScanStatusDisplay';
import ConfigurationAlertLink from './ConfigurationAlertLink';

const TabGeneral = ({data}) => {
    const {scanConfig, scanConfigSnapshot, startTime, endTime, summary, state, stateMessage, stateReason} = data || {};
    const {scope, scanFamiliesConfig} = scanConfigSnapshot;
    const {all, regions, instanceTagSelector, instanceTagExclusion} = scope;
    const {jobsCompleted, jobsLeftToRun} = summary;
 
    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <TitleValueDisplayColumn>
                    <ConfigurationAlertLink updatedConfigData={scanConfig} scanConfigData={scanConfigSnapshot} />
                    <TitleValueDisplay title="Scope"><ScopeDisplay all={all} regions={regions} /></TitleValueDisplay>
                    <TitleValueDisplay title="Included instances"><InstancesDisplay tags={instanceTagSelector}/></TitleValueDisplay>
                    <TitleValueDisplay title="Excluded instances"><InstancesDisplay tags={instanceTagExclusion}/></TitleValueDisplay>
                    <TitleValueDisplay title="Scan types"><ScanTypesDisplay scanFamiliesConfig={scanFamiliesConfig} /></TitleValueDisplay>
                </TitleValueDisplayColumn>
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium>Status</Title>
                    <ScanStatusDisplay
                        itemsCompleted={jobsCompleted}
                        itemsLeft={jobsLeftToRun}
                        state={state}
                        stateMessage={stateMessage}
                        stateReason={stateReason}
                    />
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Started">{formatDate(startTime)}</TitleValueDisplay>
                        <TitleValueDisplay title="Ended">{formatDate(endTime)}</TitleValueDisplay>
                        <TitleValueDisplay title="Duration">{calculateDuration(startTime, endTime)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>
            )}
        />
    )
}

export default TabGeneral;