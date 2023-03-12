import React from 'react';
import { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';

import './scan-actions-display.scss';

const ScanActionsDisplay = ({data}) => {
    const {id} = data;

    return (
        <div className="scan-actions-display">
            <IconWithTooltip
                tooltipId={`${id}-stop`}
                tooltipText="Stop scan"
                name={ICON_NAMES.STOP}
                onClick={event => {
                    event.stopPropagation();
                    event.preventDefault();
                    
                    console.log("stop scan");
                }}
            />
        </div>
    );
}

export default ScanActionsDisplay;