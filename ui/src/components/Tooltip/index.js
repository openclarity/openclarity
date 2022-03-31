import React from 'react';
import ReactTooltip from 'react-tooltip';

import './tooltip.scss';

const Tooltip = ({id, text, placement="top"}) => (
    <ReactTooltip
        id={id}
        className="ac-tooltip"
        effect='solid'
        place={placement}
        textColor="white"
        backgroundColor="rgba(34, 37, 41, 0.8)"
    >
        <span>{text}</span>
    </ReactTooltip>
)

export default Tooltip;

export const TooltipWrapper = ({children, className, tooltipId, tooltipText}) => (
    <React.Fragment>
        <div data-tip data-for={tooltipId} className={className}>{children}</div>
        <Tooltip id={tooltipId} text={tooltipText} />
    </React.Fragment>
)

