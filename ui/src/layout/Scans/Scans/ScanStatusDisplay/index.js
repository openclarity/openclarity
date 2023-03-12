import React from 'react';
import ProgressBar from 'components/ProgressBar';
import { findProgressStatusFromScanState } from '../utils';

import './scan-status-display.scss';

const ScanStatusDisplay = ({itemsCompleted, itemsLeft, state, stateReason, stateMessage}) => {
    const {status, errorTitle} = findProgressStatusFromScanState({state, stateReason});

    return (
        <div className="scan-status-display-wrapper">
            <ProgressBar status={status} itemsCompleted={itemsCompleted} itemsLeft={itemsLeft} />
            {errorTitle &&
                <div className="error-display-wrapper">
                    <div className="error-display-title">{errorTitle}</div>
                    <div className="error-display-message">{stateMessage || errorTitle}</div>
                </div>
            }
        </div>
    )
}

export default ScanStatusDisplay;