import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';

import './info-icon.scss';

const InfoIcon = ({tooltipId, tooltipText}) => (
    <TooltipWrapper className="info-icon-wrapper" tooltipId={tooltipId} tooltipText={tooltipText}>
        <Icon name={ICON_NAMES.INFO} />
    </TooltipWrapper>
)

export default InfoIcon;