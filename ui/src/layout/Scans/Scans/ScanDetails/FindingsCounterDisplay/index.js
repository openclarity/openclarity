import React from 'react';
import Icon from 'components/Icon';

import './findings-counter-display.scss';

const FindingsCounterDisplay = ({icon, count, title}) => (
    <div className="findings-item-display">
        <Icon name={icon} size={30} />
        <div className="findings-item-counter">{count}</div>
        <div className="findings-item-title">{title}</div>
    </div>
)

export default FindingsCounterDisplay;