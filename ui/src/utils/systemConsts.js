import COLORS from 'utils/scss_variables.module.scss';

export const ROUTES = {
    DEFAULT: "/",
    APPLICATIONS: "/applications",
    APPLICATION_RESOURCES: "/applicationResources",
    PACKAGES: "/packages",
    VULNERABILITIES: "/vulnerabilities",
    RUNTIME_SCAN: "/runtimeScan"
}

export const SEVERITY_ITEMS = {
    CRITICAL: {value: "CRITICAL", label: "Critical", color: COLORS["color-error-dark"]},
    HIGH: {value: "HIGH", label: "High", color: COLORS["color-error"]},
    MEDIUM: {value: "MEDIUM", label: "Medium", color: COLORS["color-warning"]},
    LOW: {value: "LOW", label: "Low", color: COLORS["color-warning-low"]},
    NEGLIGIBLE: {value: "NEGLIGIBLE", label: "Negligible", color: COLORS["color-status-blue"]}
};