import React from 'react';
import TitleValueDisplay, { ValuesListDisplay } from 'components/TitleValueDisplay';
import { TagsList } from 'components/Tag';
import { formatTagsToStringInstances, getEnabledScanTypesList, getScanTimeTypeTag } from 'layout/Scans/utils';
import { cronExpressionToHuman, formatDate } from 'utils/utils';
import { ScopeDisplay } from './scopeDisplayUtils';

const InstancesDisplay = ({tags}) => (
    <TagsList items={formatTagsToStringInstances(tags)} />
)

const FlagPropDisplay = ({checked, label}) => <div style={{marginBottom: "20px"}}>{`${label} ${checked ? "enabled" : "disabled"}`}</div> 

const ConfigurationReadOnlyDisplay = ({configData}) => {
    const {scope, scanFamiliesConfig, scheduled, maxParallelScanners, scannerInstanceCreationConfig} = configData;
    const {allRegions, regions, instanceTagSelector, instanceTagExclusion, shouldScanStoppedInstances} = scope;
    const {cronLine, operationTime} = scheduled;
    const {useSpotInstances} = scannerInstanceCreationConfig || {};

    return (
        <>
            <TitleValueDisplay title="Scope"><ScopeDisplay all={allRegions} regions={regions} /></TitleValueDisplay>
            <TitleValueDisplay title="Instances">
                <FlagPropDisplay label="Scan also non-running instances" checked={shouldScanStoppedInstances} />
                <TitleValueDisplay title="Included instances" isSubItem><InstancesDisplay tags={instanceTagSelector}/></TitleValueDisplay>
                <TitleValueDisplay title="Excluded instances" isSubItem><InstancesDisplay tags={instanceTagExclusion}/></TitleValueDisplay>
            </TitleValueDisplay>
            <TitleValueDisplay title="Scan types"><ValuesListDisplay values={getEnabledScanTypesList(scanFamiliesConfig)} /></TitleValueDisplay>
            <TitleValueDisplay title="Time configuration">
                <>
                    <div style={{marginBottom: "5px", fontWeight: "bold"}}>{getScanTimeTypeTag({cronLine, operationTime})}</div>
                    <div>{!!cronLine ? cronExpressionToHuman(cronLine) : formatDate(operationTime)}</div>
                </>
            </TitleValueDisplay>
            <TitleValueDisplay title="Advanced settings">
                <FlagPropDisplay label="Spot instances required" checked={useSpotInstances} />
                <TitleValueDisplay title="Maximum parallel scans" isSubItem>{maxParallelScanners}</TitleValueDisplay>
            </TitleValueDisplay>
        </>
    )
}

export default ConfigurationReadOnlyDisplay;