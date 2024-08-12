import React from 'react';
import ProgressBar, { STATUS_MAPPPING } from 'components/ProgressBar';
import ErrorMessageDisplay from 'components/ErrorMessageDisplay';

import './scan-progress-bar.scss';

export const SCAN_STATES = {
    Pending: {state: "Pending", title: "Pending"},
    Discovered: {state: "Discovered", title: "Discovered"},
    InProgress: {state: "InProgress", title: "In progress"},
    Failed: {state: "Failed", title: "Failed"},
    Done: {state: "Done", title: "Done"},
    Aborted: {state: "Aborted", title: "Aborted"},
}

const SCAN_STATES_AND_REASONS_MAPPINGS = [
    {...SCAN_STATES.Pending, reason: "Created", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {...SCAN_STATES.Discovered, reason: "AssetsDiscovered",  status: STATUS_MAPPPING.IN_PROGRESS.value},
    {...SCAN_STATES.InProgress, reason: "AssetScansRunning", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {...SCAN_STATES.Failed, reason: "Cancellation", status: STATUS_MAPPPING.STOPPED.value},
    {...SCAN_STATES.Failed, reason: "Timeout", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Scan has been timed out"},
    {...SCAN_STATES.Failed, reason: "AssetScanFailed", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Some of the elements were failed to be scanned"},
    {...SCAN_STATES.Done, reason: "NothingToScan", status: STATUS_MAPPPING.SUCCESS.value},
    {...SCAN_STATES.Done, reason: "Success", status: STATUS_MAPPPING.SUCCESS.value},
    {...SCAN_STATES.Aborted, reason: "Cancellation", status: STATUS_MAPPPING.STOPPED.value}
];

const ScanProgressBar = ({itemsCompleted, itemsLeft, state, stateReason, stateMessage, barWidth, isMinimized=false, minimizedTooltipId=null}) => {
    const {status, errorTitle} = SCAN_STATES_AND_REASONS_MAPPINGS.find(item => item.state === state && item.reason === stateReason) || {};

    return (
        <div className="scan-progres-bar-wrapper">
            <ProgressBar
                status={status}
                itemsCompleted={itemsCompleted}
                itemsLeft={itemsLeft}
                width={barWidth}
                message={isMinimized ? errorTitle : null}
                messageTooltipId={minimizedTooltipId}
            />
            {!isMinimized && errorTitle &&
                <div className="error-display-wrapper">
                    <div className="error-display-title">{errorTitle}</div>
                    <ErrorMessageDisplay>{stateMessage || errorTitle}</ErrorMessageDisplay>
                </div>
            }
        </div>
    )
}

export default ScanProgressBar;