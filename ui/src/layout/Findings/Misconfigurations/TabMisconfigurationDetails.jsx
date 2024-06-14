import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';
import { MISCONFIGURATION_SEVERITY_MAP } from './utils';

const TabMisconfigurationDetails = ({data}) => {
    const {findingInfo, firstSeen, lastSeen} = data;
    const {id, severity, description, scannerName, location, remediation, category, message} = findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="ID">{id}</TitleValueDisplay>
                        <TitleValueDisplay title="Severity">{MISCONFIGURATION_SEVERITY_MAP[severity]}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Scanner Name">{scannerName}</TitleValueDisplay>
                        <TitleValueDisplay title="File path">{location}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Category">{category}</TitleValueDisplay>
                        <TitleValueDisplay title="Remediation">{remediation}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Message" withOpen defaultOpen>{message}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Description" withOpen defaultOpen>{description}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <FindingsDetailsCommonFields firstSeen={firstSeen} lastSeen={lastSeen} />
                </>
            )}
        />
    )
}

export default TabMisconfigurationDetails;
