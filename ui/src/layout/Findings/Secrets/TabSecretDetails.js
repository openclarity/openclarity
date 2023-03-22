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
                        <TitleValueDisplay title="Start Line">{startLine}</TitleValueDisplay>
                        <TitleValueDisplay title="End line">{endLine}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="File path">{filePath}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>  
            )}
        />
    )
}

export default TabSecretDetails;