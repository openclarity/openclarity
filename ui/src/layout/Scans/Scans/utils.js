import { ICON_NAMES } from 'components/Icon';
import { STATUS_MAPPPING } from 'components/ProgressBar';

export const FINDINGS_MAPPING = {
    PACKAGES: {value: "totalPackages", title: "Packages", icon: ICON_NAMES.PACKAGE},
    EXPLOITS: {value: "totalExploits", title: "Exploits", icon: ICON_NAMES.BOMB},
    MISCONFIGURATIONS: {value: "totalMisconfigurations", title: "Misconfigurations", icon: ICON_NAMES.COG},
    SECRETS: {value: "totalSecrets", title: "Secrets", icon: ICON_NAMES.KEY},
    MALWARE: {value: "totalMalware", title: "Malwares", icon: ICON_NAMES.BUG},
    ROOTKITS: {value: "totalRootkits", title: "Rootkits", icon: ICON_NAMES.GHOST}
}

const SCAN_STATES_AND_REASONS_MAPPINGS = [
    {state: "Pending", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "Discovered", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "InProgress", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "Failed", stateReason: "Aborted", status: STATUS_MAPPPING.STOPPED.value},
    {state: "Failed", stateReason: "TimedOut", status: STATUS_MAPPPING.WARNING.value},
    {state: "Failed", stateReason: "OneOrMoreTargetFailedToScan", status: STATUS_MAPPPING.STOPPED.value, errorTitle: "Some of the elements were failed to be scanned"},
    {state: "Failed", stateReason: "DiscoveryFailed", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Discovery failed"},
    {state: "Failed", stateReason: "Unexpected", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Unexpected error occured"},
    {state: "Done", status: STATUS_MAPPPING.SUCCESS.value}
]

export const findProgressStatusFromScanState = ({state, stateReason}) => (
    SCAN_STATES_AND_REASONS_MAPPINGS.find(item => item.state === state && (!item.stateReason || item.stateReason === stateReason)) || {}
)