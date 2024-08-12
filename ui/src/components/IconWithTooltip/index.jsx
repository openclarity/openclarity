import React from 'react';
import Icon from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';

const IconWithTooltip = ({tooltipId, tooltipText, size=22, ...props}) => (
    <div className="icon-with-tooltip-wrapper" style={{height: `${size}px`, width: `${size}px`}}>
        <TooltipWrapper tooltipId={tooltipId} tooltipText={tooltipText}>
            <Icon {...props} size={size} />
        </TooltipWrapper>
    </div>
)

export default IconWithTooltip;