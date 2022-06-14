import React from 'react';
import TitleValueDisplay, { TitleValueDisplayColumn } from 'components/TitleValueDisplay';
import { CisBenchmarkLevelsDisplay } from 'components/VulnerabilitiesSummaryDisplay';
import { getItemsString } from 'utils/utils';
import { VulnerabilitiesLink, PackagesLink, ApplicationsLink } from '../utils';

const TabDetails = ({data}) => {
    const {applicationResource, licenses} = data || {};
    const {id, resourceName, resourceHash, resourceType, vulnerabilities, applications, packages, reportingSBOMAnalyzers, cisDockerBenchmarkResults} = applicationResource || {};
    
    return (
        <div className="application-resource-tab-details">
            <TitleValueDisplayColumn>
                <TitleValueDisplay title="Resource name">{resourceName}</TitleValueDisplay>
                <TitleValueDisplay title="Resource Hash">{resourceHash}</TitleValueDisplay>
                <TitleValueDisplay title="Resource Type">{resourceType}</TitleValueDisplay>
                <TitleValueDisplay title="Licenses">{getItemsString(licenses)}</TitleValueDisplay>
                <TitleValueDisplay title="SBOM Analyzers">{getItemsString(reportingSBOMAnalyzers)}</TitleValueDisplay>
            </TitleValueDisplayColumn>
            <TitleValueDisplayColumn>
                <TitleValueDisplay title="Applications">
                    <ApplicationsLink applications={applications} applicationResourceID={id} resourceName={resourceName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Packages">
                    <PackagesLink packages={packages} applicationResourceID={id} resourceName={resourceName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Vulnerabilities">
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationResourceID={id} resourceName={resourceName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="CIS Docker Benchmark">
                    <CisBenchmarkLevelsDisplay id={id} levels={cisDockerBenchmarkResults} withTotal />
                </TitleValueDisplay>
            </TitleValueDisplayColumn>
        </div>
    )
}

export default TabDetails;