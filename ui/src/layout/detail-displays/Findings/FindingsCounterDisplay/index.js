import React from 'react';
import Icon from 'components/Icon';
import { formatNumber } from 'utils/utils';

import './findings-counter-display.scss';

const FindingsCounterDisplay = ({icon, color, count, title}) => (
    <div className="findings-item-display">
        <Icon name={icon} size={30} style={{color}} />
        <div className="findings-item-counter">{formatNumber(count)}</div>
        <div className="findings-item-title">{title}</div>
    </div>
)

export default FindingsCounterDisplay;