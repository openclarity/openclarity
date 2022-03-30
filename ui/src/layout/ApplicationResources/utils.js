import { PackagesLink as GeneralPackagesLink } from 'layout/Packages';
import { VulnerabilitiesLink as GeneralVulnerabilitiesLink } from 'layout/Vulnerabilities';
import { ApplicationsLink as GeneralApplicationsLink } from 'layout/Applications';
import { BoldText } from 'utils/utils';

const getTitle = name => <span>{`resource: `}<BoldText>{name}</BoldText></span>;

export const VulnerabilitiesLink = ({id, applicationResourceID, vulnerabilities, resourceName}) => (
    <GeneralVulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)

export const PackagesLink = ({applicationResourceID, packages, resourceName}) => (
    <GeneralPackagesLink value={packages} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)

export const ApplicationsLink = ({applicationResourceID, applications, resourceName}) => (
    <GeneralApplicationsLink count={applications} applicationResourceID={applicationResourceID} title={getTitle(resourceName)} />
)