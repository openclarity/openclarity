import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';

import COLORS from 'utils/scss_variables.module.scss';

import './progress-bar.scss';

export const STATUS_MAPPPING = {
    IN_PROGRESS: {value: "IN_PROGRESS", color: COLORS["color-main"]},
    SUCCESS: {value: "SUCCESS", icon: ICON_NAMES.CHECK_MARK, color: COLORS["color-success"]},
    ERROR: {value: "ERROR", icon: ICON_NAMES.X_MARK, color: COLORS["color-error"]},
    STOPPED: {value: "STOPPED", icon: ICON_NAMES.BLOCK, color: COLORS["color-grey"]},
    WARNING: {value: "WARNING", icon: ICON_NAMES.WARNING, color: COLORS["color-success"], iconColor: COLORS["color-warning"]}
}

const ProgressBar = ({status=STATUS_MAPPPING.IN_PROGRESS.value, itemsCompleted=0, itemsLeft=0, width="100%", message=null, messageTooltipId=null, customeTitle}) => {
    const totalItems = itemsCompleted + itemsLeft;
    const percent = status === STATUS_MAPPPING.IN_PROGRESS.value ? (!!totalItems ? Math.round((itemsCompleted / totalItems) * 100) : 0) : 100;

    const {icon, color, iconColor} = STATUS_MAPPPING[status];
    const progressIconColor = iconColor || color;
    const IconComponent = !!message ? IconWithTooltip : Icon;

    return (
        <div className="progress-bar-wrapper">
            <div className="progress-bar-container" style={{width}}>
                <div className={classnames("progress-bar-filler", {done: percent === 100})} style={{width: `${percent}%`, backgroundColor: color}}></div>  
            </div>
            {!icon ? <div className="progress-bar-title">{!!customeTitle ? customeTitle : `${percent}%`}</div> :
                <IconComponent
                    name={icon}
                    style={{color: progressIconColor}}
                    tooltipId={`progress-bar-message-${messageTooltipId}`}
                    tooltipText={message}
                />
            }
        </div>
    )
}

export default ProgressBar;