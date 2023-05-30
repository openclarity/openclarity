import { PackagesLink as GeneralPackagesLink } from 'layout/Packages';
import { VulnerabilitiesLink as GeneralVulnerabilitiesLink } from 'layout/Vulnerabilities';
import { ApplicationsLink as GeneralApplicationsLink } from 'layout/Applications';

export const RESOURCE_TYPES = {
    IMAGE: {value: "IMAGE", label: "Image"},
    DIRECTORY: {value: "DIRECTORY", label: "Directory"},
    FILE: {value: "FILE", label: "File"}
};

const getTitle = name => `resource: ${name}`;

export const VulnerabilitiesLink = ({id, applicationResourceID, vulnerabilities, resourceName}) => (
    <GeneralVulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)

export const PackagesLink = ({applicationResourceID, packages, resourceName}) => (
    <GeneralPackagesLink value={packages} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)

export const ApplicationsLink = ({applicationResourceID, applications, resourceName}) => (
    <GeneralApplicationsLink count={applications} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)