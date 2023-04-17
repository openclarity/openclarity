import React from 'react';
import { uniqBy } from 'lodash';
import ProgressBar, { STATUS_MAPPPING } from 'components/ProgressBar';
import ErrorMessageDisplay from 'components/ErrorMessageDisplay';

import './scan-progress-bar.scss';

const SCAN_STATES_AND_REASONS_MAPPINGS = [
    {state: "Pending", title: "Pending", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "Discovered", title: "Discovered", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "InProgress", title: "In progress", status: STATUS_MAPPPING.IN_PROGRESS.value},
    {state: "Failed", title: "Failed", stateReason: "Aborted", status: STATUS_MAPPPING.STOPPED.value},
    {state: "Failed", title: "Failed", stateReason: "TimedOut", status: STATUS_MAPPPING.WARNING.value},
    {state: "Failed", title: "Failed", stateReason: "OneOrMoreTargetFailedToScan", status: STATUS_MAPPPING.WARNING.value, errorTitle: "Some of the elements were failed to be scanned"},
    {state: "Failed", title: "Failed", stateReason: "DiscoveryFailed", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Discovery failed"},
    {state: "Failed", title: "Failed", stateReason: "Unexpected", status: STATUS_MAPPPING.ERROR.value, errorTitle: "Unexpected error occured"},
    {state: "Done", title: "Done", status: STATUS_MAPPPING.SUCCESS.value}
];

export const SCAN_STATUS_ITEMS = uniqBy(SCAN_STATES_AND_REASONS_MAPPINGS, "state").map(({state, title}) => ({value: state, label: title}))

const ScanProgressBar = ({itemsCompleted, itemsLeft, state, stateReason, stateMessage, barWidth, isMinimized=false, minimizedTooltipId=null}) => {
    const {status, errorTitle} = SCAN_STATES_AND_REASONS_MAPPINGS
        .find(item => item.state === state && (!item.stateReason || item.stateReason === stateReason)) || {};

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