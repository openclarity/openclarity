import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM, ROUTES } from 'utils/systemConsts';
import Vulnerabilities from './Vulnerabilities';
import Exploits from './Exploits';
import Misconfigurations from './Misconfigurations';
import Secrets from './Secrets';
import Malware from './Malware';
import Rootkits from './Rootkits';
import Packages from './Packages';

const FINDINGS_TAB_ITEMS = {
    VULNERABILITIES: {
        id: "vulnerabilities",
        path: "vulnerabilities",
        title: "Vulnerabilities",
        component: Vulnerabilities,
        findingsType: VULNERABIITY_FINDINGS_ITEM.value
    },
    EXPLOITS: {
        id: "exploits",
        path: "exploits",
        title: "Exploits",
        component: Exploits,
        findingsType: FINDINGS_MAPPING.EXPLOITS.value
    },
    MISCONFIGURATIONS: {
        id: "misconfigurations",
        path: "misconfigurations",
        title: "Misconfigurations",
        component: Misconfigurations,
        findingsType: FINDINGS_MAPPING.MISCONFIGURATIONS.value
    },
    SECRETS: {
        id: "secrets",
        path: "secrets",
        title: "Secrets",
        component: Secrets,
        findingsType: FINDINGS_MAPPING.SECRETS.value
    },
    MALWARE: {
        id: "malware",
        path: "malware",
        title: "Malware",
        component: Malware,
        findingsType: FINDINGS_MAPPING.MALWARE.value
    },
    ROOTKITS: {
        id: "rootkits",
        path: "rootkits",
        title: "Rootkits",
        component: Rootkits,
        findingsType: FINDINGS_MAPPING.ROOTKITS.value
    },
    PACKAGES: {
        id: "packages",
        path: "packages",
        title: "Packages",
        component: Packages,
        findingsType: FINDINGS_MAPPING.PACKAGES.value
    }
}

export const FINDINGS_PATHS = Object.keys(FINDINGS_TAB_ITEMS).reduce((acc, curr) => ({...acc, [curr]: FINDINGS_TAB_ITEMS[curr].path}), {});

export const getFindingsAbsolutePathByFindingType = type => {
    const relativePath = Object.values(FINDINGS_TAB_ITEMS).find(({findingsType}) => type === findingsType)?.path;

    return `${ROUTES.FINDINGS}/${relativePath}`
};

const Findings = () => {
    const {pathname} = useLocation();
    
    return (
        <TabbedPage
            redirectTo={`${pathname}/${FINDINGS_PATHS.VULNERABILITIES}`}
            items={Object.values(FINDINGS_TAB_ITEMS)}
            withStickyTabs
        />
    )
}

export default Findings;
