import { PackagesLink as GeneralPackagesLink } from 'layout/Packages';
import { VulnerabilitiesLink as GeneralVulnerabilitiesLink } from 'layout/Vulnerabilities';
import { ApplicationResourcesLink as GeneralApplicationResourcesLink } from 'layout/ApplicationResources';

export const APPLICATION_TYPE_ITEMS = [
    {value: "POD", label: "Pod"},
    {value: "DIRECTORY", label: "Directory"},
    {value: "LAMBDA", label: "Lambda"}
];

const getTitle = name => `application: ${name}`;

export const VulnerabilitiesLink = ({id, applicationID, vulnerabilities, applicationName}) => (
    <GeneralVulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationID={applicationID} title={getTitle(applicationName)} />
)

export const PackagesLink = ({applicationID, packages, applicationName}) => (
    <GeneralPackagesLink value={packages} applicationID={applicationID} title={getTitle(applicationName)} />
)

export const ApplicationResourcesLink = ({applicationID, applicationResources, applicationName}) => (
    <GeneralApplicationResourcesLink count={applicationResources} applicationID={applicationID} title={getTitle(applicationName)} />
)