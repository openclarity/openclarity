import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';

const TabSecretDetails = ({data}) => {
    const {fingerprint, description, startLine, endLine, filePath} = data.findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Fingerprint">{fingerprint}</TitleValueDisplay>
                        <TitleValueDisplay title="Description">{description}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="StartLine">{startLine}</TitleValueDisplay>
                        <TitleValueDisplay title="EndLine">{endLine}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="FilePath">{filePath}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>  
            )}
        />
    )
}

export default TabSecretDetails;