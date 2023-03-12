import React from 'react';
import TabbedPage from 'components/TabbedPage';

const FINDINGS_PATHS = {
    VULNERABILITIES: "vulnerabilities",
    EXPLOITS: "exploits",
    MISCONFIGURATIONS: "misconfigurations",
    SECRETS: "secrets",
    MALWARE: "malware",
    ROOTKITS: "rootkits",
    PACKAGES: "packages"
}

const Findings = () => (
    <TabbedPage
        items={[
            {
                id: "vulnerabilities",
                title: "Vulnerabilities",
                isIndex: true,
                component: () => "TBD"
            },
            {
                id: "exploits",
                title: "Exploits",
                path: FINDINGS_PATHS.EXPLOITS,
                component: () => "TBD"
            },
            {
                id: "misconfigurations",
                title: "misconfigurations",
                path: FINDINGS_PATHS.MISCONFIGURATIONS,
                component: () => "TBD"
            },
            {
                id: "secrets",
                title: "Secrets",
                path: FINDINGS_PATHS.SECRETS,
                component: () => "TBD"
            },
            {
                id: "malware",
                title: "Malware",
                path: FINDINGS_PATHS.MALWARE,
                component: () => "TBD"
            },
            {
                id: "rootkits",
                title: "Rootkits",
                path: FINDINGS_PATHS.ROOTKITS,
                component: () => "TBD"
            },
            {
                id: "packages",
                title: "Packages",
                path: FINDINGS_PATHS.PACKAGES,
                component: () => "TBD"
            }
        ]}
        withStickyTabs
    />
)

export default Findings;