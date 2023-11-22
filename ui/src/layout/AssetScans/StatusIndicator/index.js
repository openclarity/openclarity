import React from 'react';
import { isEmpty } from 'lodash';
import { TooltipWrapper } from 'components/Tooltip';

import COLORS from 'utils/scss_variables.module.scss';

import './status-indicator.scss';

export const STATUS_MAPPING = {
    Pending: {title: "Pending", color: COLORS["color-main"]},
    InProgress: {title: "In Progress", color: COLORS["color-main"]},
    Skipped: {title: "Skipped", color: COLORS["color-grey"]},
    Failed: {title: "Failed", color: COLORS["color-error"]},
    Done: {title: "Done", color: COLORS["color-success"]},

    NotScanned: {title: "Not Scanned", color: COLORS["color-grey"]},
    Scheduled: {title: "Scheduled", color: COLORS["color-main"]},
    ReadyToScan: {title: "Ready To Scan", color: COLORS["color-main"]},
    Aborted: {title: "Aborted", color: COLORS["color-grey"]}
}

const StatusIndicator = ({state, isError=false}) => {
    const {title, color} = STATUS_MAPPING[state] || {};

    return (
        <div className="status-indicator-wrapper">
            <div className="status-indicator" style={{backgroundColor: isError ? COLORS["color-warning"] : color}}></div>
            <div className="status-title">{title}</div>
        </div>
    )
}

const StatusIndicatorWrapper = ({state, errors, tooltipId}) => (
    <div>
        {
            (!isEmpty(errors) && !isEmpty(tooltipId)) ? (
                <TooltipWrapper tooltipId={`status-inicator-${tooltipId}`} tooltipText="An error has occurred">
                    <StatusIndicator state={state} isError={true} />
                </TooltipWrapper>
            ) : <StatusIndicator state={state} isError={!isEmpty(errors)} />
        }
    </div>
    
);

export default StatusIndicatorWrapper;
