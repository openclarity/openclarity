import React from 'react';
import PageContainer from 'components/PageContainer';
import TitleValueDisplay, { TitleValueDisplayColumn } from 'components/TitleValueDisplay';
import DetailsPageWrapper from 'components/DetailsPageWrapper';
import { LabelsDisplay } from 'components/LabelTag';
import { getItemsString } from 'utils/utils';
import { VulnerabilitiesLink, PackagesLink, ApplicationResourcesLink } from './utils';

const DetailsContent = ({data}) => {
    const {application, licenses} = data || {};
    const {id, applicationName, applicationType, labels, environments, applicationResources, packages, vulnerabilities} = application || {};

    return (
        <PageContainer className="application-details-container" withPadding>
            <TitleValueDisplayColumn>
                <TitleValueDisplay title="Name">{applicationName}</TitleValueDisplay>
                <TitleValueDisplay title="Type">{applicationType}</TitleValueDisplay>
                <TitleValueDisplay title="ID">{id}</TitleValueDisplay>
            </TitleValueDisplayColumn>
            <TitleValueDisplayColumn>
                <TitleValueDisplay title="Environments">{getItemsString(environments)}</TitleValueDisplay>
                <TitleValueDisplay title="Licenses">{getItemsString(licenses)}</TitleValueDisplay>
                <TitleValueDisplay title="Labels"><LabelsDisplay labels={labels} wrapLabels /></TitleValueDisplay>
            </TitleValueDisplayColumn>
            <TitleValueDisplayColumn>
                <TitleValueDisplay title="Number of resources">
                    <ApplicationResourcesLink applicationResources={applicationResources} applicationID={id} applicationName={applicationName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Number of packages">
                    <PackagesLink packages={packages} applicationID={id} applicationName={applicationName} />
                </TitleValueDisplay>
                <TitleValueDisplay title="Vulnerabilities">
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationID={id} applicationName={applicationName} />
                </TitleValueDisplay>
            </TitleValueDisplayColumn>
        </PageContainer>
    )
}

const ApplicationDetails = () => (
    <DetailsPageWrapper title="Application information" backTitle="Applications" url="applications" detailsContent={DetailsContent} />
)

export default ApplicationDetails;