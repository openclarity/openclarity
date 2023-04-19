import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';

const TabSecretDetails = ({data}) => {
    const {findingInfo, foundOn, invalidatedOn} = data;
    const {fingerprint, description, startLine, endLine, filePath} = findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Fingerprint">{fingerprint}</TitleValueDisplay>
                        <TitleValueDisplay title="Description">{description}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Start Line">{startLine}</TitleValueDisplay>
                        <TitleValueDisplay title="End line">{endLine}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="File path">{filePath}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <FindingsDetailsCommonFields foundOn={foundOn} invalidatedOn={invalidatedOn} />
                </>  
            )}
        />
    )
}

export default TabSecretDetails;