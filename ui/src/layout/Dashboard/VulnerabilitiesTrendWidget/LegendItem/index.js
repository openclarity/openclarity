import React from 'react';
import classnames from 'classnames';

import './legend-item.scss';

const LegendItem = ({title, color, isDarkMode=false}) => (
    <div className={classnames("trend-legend-item", {"dark-mode": isDarkMode})}>
        <div className="trend-legend-item-indicatior" style={{backgroundColor: color}}></div>
        <div className="trend-legend-item-title">{title}</div>
    </div>
);

export default LegendItem;