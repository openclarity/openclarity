import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow, ValuesListDisplay } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';

const TabPackageDetails = ({data}) => {
    const {findingInfo, foundOn, invalidatedOn} = data;
    const {name, version, language, licenses} = findingInfo;

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Package name">{name}</TitleValueDisplay>
                        <TitleValueDisplay title="Version">{version}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Language">{language}</TitleValueDisplay>
                        <TitleValueDisplay title="Licenses"><ValuesListDisplay values={licenses} /></TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <FindingsDetailsCommonFields foundOn={foundOn} invalidatedOn={invalidatedOn} />
                </>  
            )}
        />
    )
}

export default TabPackageDetails;
