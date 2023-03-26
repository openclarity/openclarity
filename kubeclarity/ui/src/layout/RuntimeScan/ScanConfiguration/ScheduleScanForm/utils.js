import { reduce } from 'lodash';

export const SCHEDULE_TYPES = {
    LATER: {value: "SingleScheduleScanConfig", label: "Later"},
    REPETITIVE: {value: "REPETITIVE", label: "Repetitive"}
};

export const GENERAL_FOMR_FIELDS = {
    SCHEDULE_TYPE: "ScheduleScanConfigType",
    NAMESPACES: "namespaces",
    CIS_ENABLED: "cisDockerBenchmarkScanEnabled",
    MAX_SCANPARALLELISM: "maxScanParallelism"
}

export const SCHEDULE_TYPE_DATA_WRAPPER = "scanConfigType";

export const formatFormFields = formFields => reduce(formFields, (prev, curr, key) => ({...prev, [key]: `${SCHEDULE_TYPE_DATA_WRAPPER}.${curr}`}), {});