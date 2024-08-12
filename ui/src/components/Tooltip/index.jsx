import React from 'react';
import classnames from 'classnames';
import ReactTooltip from 'react-tooltip';

import './tooltip.scss';

const Tooltip = ({id, text, placement="top"}) => (
    <ReactTooltip
        id={id}
        className="clarity-tooltip"
        effect='solid'
        place={placement}
        textColor="white"
        backgroundColor="rgba(34, 37, 41, 1)"
    >
        <span>{text}</span>
    </ReactTooltip>
)

export default Tooltip;

export const TooltipWrapper = ({children, className, tooltipId, tooltipText}) => (
    <React.Fragment>
        <div data-tip data-for={tooltipId} className={classnames("tooltip-wrapper", className)} style={{display: "inline-block"}}>{children}</div>
        <Tooltip id={tooltipId} text={tooltipText} />
    </React.Fragment>
)