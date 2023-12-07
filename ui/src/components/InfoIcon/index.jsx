import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';

import './info-icon.scss';

const InfoIcon = ({tooltipId, tooltipText, large=false}) => (
    <TooltipWrapper className={classnames("info-icon-wrapper", {large})} tooltipId={tooltipId} tooltipText={tooltipText}>
        <Icon name={ICON_NAMES.INFO} size={large ? 12 : 10} />
    </TooltipWrapper>
)

export default InfoIcon;