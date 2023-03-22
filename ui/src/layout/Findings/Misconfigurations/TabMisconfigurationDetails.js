import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';

const TabMisconfigurationDetails = ({data}) => {
    const {path, description} = data.findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="File path">{path}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Description" withOpen defaultOpen>{description}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>  
            )}
        />
    )
}

export default TabMisconfigurationDetails;