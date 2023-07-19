import React from 'react';
import TitleValueDisplay, { ValuesListDisplay, TitleValueDisplayRow } from 'components/TitleValueDisplay';
import Title from 'components/Title';
import { getEnabledScanTypesList, getScanTimeTypeTag } from 'layout/Scans/utils';
import { cronExpressionToHuman, formatDate } from 'utils/utils';

const FlagPropDisplay = ({checked, label}) => <div style={{marginBottom: "20px"}}>{`${label} ${checked ? "enabled" : "disabled"}`}</div> 

const ConfigurationReadOnlyDisplay = ({configData}) => {
    const {scanTemplate, scheduled} = configData;
    const {scope, maxParallelScanners, assetScanTemplate} = scanTemplate;
    const {scannerInstanceCreationConfig, scanFamiliesConfig} = assetScanTemplate;
    const {cronLine, operationTime} = scheduled;
    const {useSpotInstances} = scannerInstanceCreationConfig || {};

    return (
        <>
            <Title medium>Schedule</Title>
            <TitleValueDisplayRow>
                <>
                    <div style={{marginBottom: "5px", fontWeight: "bold"}}>{getScanTimeTypeTag({cronLine, operationTime})}</div>
                    <div>{!!cronLine ? cronExpressionToHuman(cronLine) : formatDate(operationTime)}</div>
                </>
            </TitleValueDisplayRow>
            <Title medium>Scan Configuration</Title>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Scope">{scope}</TitleValueDisplay>
                <TitleValueDisplay title="Maximum parallel scans">{maxParallelScanners}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <Title medium>Asset Scan Configuration</Title>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Scan types"><ValuesListDisplay values={getEnabledScanTypesList(scanFamiliesConfig)} /></TitleValueDisplay>
                <TitleValueDisplay title="Advanced settings">
                    <FlagPropDisplay label="Spot instances required" checked={useSpotInstances} />
                </TitleValueDisplay>
            </TitleValueDisplayRow>
        </>
    )
}

export const ScanReadOnlyDisplay = ({scanData}) => {
    const {scope, maxParallelScanners, assetScanTemplate} = scanData;
    const {scannerInstanceCreationConfig, scanFamiliesConfig} = assetScanTemplate;
    const {useSpotInstances} = scannerInstanceCreationConfig || {};

    return (
        <>
            <Title medium>Scan Configuration</Title>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Scope">{scope}</TitleValueDisplay>
                <TitleValueDisplay title="Maximum parallel scans">{maxParallelScanners}</TitleValueDisplay>
            </TitleValueDisplayRow>
            <Title medium>Asset Scan Configuration</Title>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Scan types"><ValuesListDisplay values={getEnabledScanTypesList(scanFamiliesConfig)} /></TitleValueDisplay>
                <TitleValueDisplay title="Advanced settings">
                    <FlagPropDisplay label="Spot instances required" checked={useSpotInstances} />
                </TitleValueDisplay>
            </TitleValueDisplayRow>
        </>
    )
}

export default ConfigurationReadOnlyDisplay;
