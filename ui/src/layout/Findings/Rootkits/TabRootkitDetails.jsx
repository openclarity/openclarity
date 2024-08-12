import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';

const TabPackageDetails = ({data}) => {
    const {findingInfo, foundOn, invalidatedOn} = data;
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
                    <FindingsDetailsCommonFields foundOn={foundOn} invalidatedOn={invalidatedOn} />
                </>  
            )}
        />
    )
}

export default TabPackageDetails;
