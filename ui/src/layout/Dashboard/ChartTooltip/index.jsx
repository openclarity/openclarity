import React from 'react';
import Icon from 'components/Icon';
import { BoldText, formatNumber } from 'utils/utils';

import './chart-tooltip.scss';

const TooltipCountItem = ({color, icon, value, title}) => (
    <div className="widget-chart-content-item">
        <Icon name={icon} size={18} style={{color}} />
        <div className="widget-chart-content-item-text"><BoldText>{formatNumber(value)}</BoldText>{` ${title}`}</div>
    </div>
)

const ChartTooltip = ({active, payload, widgetFindings, headerDisplay: HeaderDisplay, countKeyName}) => {
    if (active && payload && payload.length) {
        const data = payload[0].payload;

        return (
            <div className="widget-chart-tooltip">
                <div className="widget-chart-tooltip-header">
                    <HeaderDisplay data={data} />
                </div>
                <div className="widget-chart-tooltip-content">
                    {
                        widgetFindings?.map(findingMap => {
                            const {dataKey, title, icon, color, darkColor} = findingMap;
                            const countKey = findingMap[countKeyName];
                            const value = data[countKey] || 0;
                            
                            if (value === 0) {
                                return null;
                            }
                            
                            return (
                                <TooltipCountItem key={dataKey} title={title} icon={icon} value={value} color={darkColor || color} />
                            )
                        })
                    }
                </div>
            </div>
        )
    }

    return null;
}

export default ChartTooltip;