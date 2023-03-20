import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import Vulnerabilities from './Vulnerabilities';
import Exploits from './Exploits';
import Misconfigurations from './Misconfigurations';
import Secrets from './Secrets';
import Malware from './Malware';
import Rootkits from './Rootkits';
import Packages from './Packages';

const FINDINGS_TAB_ITEMS = {
    VULNERABILITIES: {id: "vulnerabilities", path: "vulnerabilities", title: "Vulnerabilities", component: Vulnerabilities},
    EXPLOITS: {id: "exploits", path: "exploits", title: "Exploits", component: Exploits},
    MISCONFIGURATIONS: {id: "misconfigurations", path: "misconfigurations", title: "Misconfigurations", component: Misconfigurations},
    SECRETS: {id: "secrets", path: "secrets", title: "Secrets", component: Secrets},
    MALWARE: {id: "malware", path: "malware", title: "Malware", component: Malware},
    ROOTKITS: {id: "rootkits", path: "rootkits", title: "Rootkits", component: Rootkits},
    PACKAGES: {id: "packages", path: "packages", title: "Packages", component: Packages},
}

export const FINDINGS_PATHS = Object.keys(FINDINGS_TAB_ITEMS).reduce((acc, curr) => ({...acc, [curr]: FINDINGS_TAB_ITEMS[curr].path}), {});

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