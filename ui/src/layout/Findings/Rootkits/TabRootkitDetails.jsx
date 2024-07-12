import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';
import AssetCountDisplay from '../AssetCountDisplay';

const TabPackageDetails = ({data}) => {
    const {id, findingInfo, firstSeen, lastSeen} = data;
    const {rootkitName, message} = findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Rootkit name">{rootkitName}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Message">{message}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <FindingsDetailsCommonFields firstSeen={firstSeen} lastSeen={lastSeen} />
                    {AssetCountDisplay(id)}
                </>
            )}
        />
    )
}

export default TabPackageDetails;
