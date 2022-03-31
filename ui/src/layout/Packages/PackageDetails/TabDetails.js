import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { VulnerabilitiesLink, ApplicationResourcesLink, ApplicationsLink } from '../utils';

const TabDetails = ({data}) => {
    const {id, packageName, version, language, license, applications, applicationResources, vulnerabilities} = data || {};

    return (
        <div className="package-tab-details">
            <div className="package-details-section">
                <TitleValueDisplayRow>
                    <TitleValueDisplay title="Package name">{packageName}</TitleValueDisplay>
                    <TitleValueDisplay title="License">{license}</TitleValueDisplay>
                </TitleValueDisplayRow>
                <TitleValueDisplayRow>
                    <TitleValueDisplay title="Version">{version}</TitleValueDisplay>
                    <TitleValueDisplay title="Language">{language}</TitleValueDisplay>
                </TitleValueDisplayRow>
            </div>
            <div className="package-dependencies-section">
                <TitleValueDisplay title="Vulnerabilities">
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} packageVersion={version} packageName={packageName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Applications">
                    <ApplicationsLink applications={applications} packageID={id} packageName={packageName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Application resources">
                    <ApplicationResourcesLink applicationResources={applicationResources} packageID={id} packageName={packageName} />
                </TitleValueDisplay>
            </div>
        </div>
    )
}

export default TabDetails;