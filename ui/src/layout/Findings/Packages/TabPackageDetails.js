import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow, ValuesListDisplay } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';

const TabPackageDetails = ({data}) => {
    const {name, version, language, licenses} = data.findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Package name">{name}</TitleValueDisplay>
                        <TitleValueDisplay title="Version">{version}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Languege">{language}</TitleValueDisplay>
                        <TitleValueDisplay title="Licenses"><ValuesListDisplay values={licenses} /></TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>  
            )}
        />
    )
}

export default TabPackageDetails;