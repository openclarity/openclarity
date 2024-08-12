import { ICON_NAMES } from 'components/Icon';

import COLORS from 'utils/scss_variables.module.scss';

export const ROUTES = {
    DEFAULT: "/",
    SCANS: "/scans",
    ASSETS: "/assets",
    ASSET_SCANS: "/assets-scans",
    FINDINGS: "/findings"
}

export const APIS = {
    SCANS: "scans",
    SCAN_CONFIGS: "scanConfigs",
    ASSETS: "assets",
    ASSET_SCANS: "assetScans",
    FINDINGS: "findings",
    DASHBOARD_RISKIEST_REGIONS: "dashboard/riskiestRegions",
    DASHBOARD_RISKIEST_ASSETS: "dashboard/riskiestAssets",
    DASHBOARD_FINDINGS_TRENDS: "dashboard/findingsTrends",
    DASHBOARD_FINDINGS_IMPACT: "dashboard/findingsImpact"
}

export const FINDINGS_MAPPING = {
    EXPLOITS: {
        value: "EXPLOITS",
        dataKey: "exploits",
        totalKey: "totalExploits",
        typeKey: "EXPLOIT",
        title: "Exploits",
        icon: ICON_NAMES.BOMB,
        color: COLORS["color-main"],
        darkColor: COLORS["color-main-variation-dark"]
    },
    MISCONFIGURATIONS: {
        value: "MISCONFIGURATIONS",
        dataKey: "misconfigurations",
        totalKey: "totalMisconfigurations",
        typeKey: "MISCONFIGURATION",
        title: "Misconfigurations",
        icon: ICON_NAMES.COG,
        color: COLORS["color-findings-1"],
        darkColor: COLORS["color-findings-1-variation-dark"]
    },
    SECRETS: {
        value: "SECRETS",
        dataKey: "secrets",
        totalKey: "totalSecrets",
        typeKey: "SECRET",
        title: "Secrets",
        icon: ICON_NAMES.KEY,
        color: COLORS["color-findings-2"]
    },
    MALWARE: {
        value: "MALWARE",
        dataKey: "malware",
        totalKey: "totalMalware",
        typeKey: "MALWARE",
        title: "Malware",
        icon: ICON_NAMES.BUG,
        color: COLORS["color-findings-3"]
    },
    ROOTKITS: {
        value: "ROOTKITS",
        dataKey: "rootkits",
        totalKey: "totalRootkits",
        typeKey: "ROOTKIT",
        title: "Rootkits",
        icon: ICON_NAMES.GHOST,
        color: COLORS["color-findings-4"]
    },
    PACKAGES: {
        value: "PACKAGES",
        dataKey:"packages",
        totalKey: "totalPackages",
        typeKey: "PACKAGE",
        title: "Packages",
        icon: ICON_NAMES.PACKAGE,
        color: COLORS["color-findings-5"]
    }
}

export const VULNERABIITY_FINDINGS_ITEM = {
    value: "VULNERABIITIES",
    dataKey: "vulnerabilities",
    typeKey: "VULNERABILITY",
    title: "Vulnerabilities",
    icon: ICON_NAMES.SHIELD,
    color: COLORS["color-main-dark"],
    darkColor: COLORS["color-main-dark-variation-dark"],
}
