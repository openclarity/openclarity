import React from 'react';
import PageContainer from 'components/PageContainer';
import TitleValueDisplay, { TitleValueDisplayColumn } from 'components/TitleValueDisplay';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import { CisBenchmarkLevelsDisplay } from 'components/VulnerabilitiesSummaryDisplay';
import { getItemsString } from 'utils/utils';
import { VulnerabilitiesLink, PackagesLink, ApplicationsLink } from './utils';

const DetailsContent = ({data}) => {
    const {applicationResource, licenses} = data || {};
    const {id, resourceName, resourceHash, resourceType, vulnerabilities, applications, packages, reportingSBOMAnalyzers, cisDockerBenchmarkResults} = applicationResource || {};
    
    return (
        <PageContainer className="application-resource-details-container" withPadding>
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
                <TitleValueDisplay title="CIS Benchmark">
                    <CisBenchmarkLevelsDisplay id={id} levels={cisDockerBenchmarkResults} withTotal />
                </TitleValueDisplay>
            </TitleValueDisplayColumn>
        </PageContainer>
    )
}

const ApplicationResourcesDetails = () => (
    <DetailsPageWrapper title="Application resource information" backTitle="Application resources" url="applicationResources" detailsContent={DetailsContent} />
)

export default ApplicationResourcesDetails;