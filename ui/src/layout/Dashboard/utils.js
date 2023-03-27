import React from 'react';
import Icon from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';

export const LegendIcon = ({widgetName, title, icon, color, size}) => (
    <TooltipWrapper tooltipId={`${widgetName}-${title}`} tooltipText={title}>
        <Icon name={icon} style={{color}} size={size} />
    </TooltipWrapper>
);