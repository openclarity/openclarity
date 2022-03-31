import React from 'react';
import { isEmpty } from 'lodash';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { formatDateBy } from 'utils/utils';
import { SEVERITY_ITEMS } from 'utils/systemConsts';
import { NO_DATA } from 'layout/Dashboard/utils';
import LegendItem from '../LegendItem';

import COLORS from 'utils/scss_variables.module.scss';

import './trends-chart.scss';

const CHART_MARGIN = 20;
const AXIS_TICK_STYLE = {fill: COLORS["color-grey-dark"], fontSize: "12px"};

const TrendsChart = ({data, timeFormat="HH:mm:ss"}) => (
    <ResponsiveContainer width="100%" height={220}>
        <AreaChart
            data={data}
            margin={{top: CHART_MARGIN, right: CHART_MARGIN, left: -10, bottom: CHART_MARGIN}}
        >
            {
                Object.values(SEVERITY_ITEMS).reverse().map(({value, label, color}) => (
                    <Area
                        key={value}
                        type="monotone"
                        dataKey={value}
                        stroke={color}
                        fill={color}
                        fillOpacity={1}
                        activeDot={{stroke: "white", strokeWidth: 2, fill: COLORS["color-grey-black"], r: 6}}
                    />
                ))
            }
            <CartesianGrid stroke="rgba(220, 220, 220, 0.26)" vertical={false} />
            <XAxis dataKey="time" stroke={COLORS["color-grey-light"]} tick={AXIS_TICK_STYLE} tickFormatter={time => time === "auto" ? NO_DATA : formatDateBy(time, timeFormat)} />
            <YAxis stroke="transparent" tick={AXIS_TICK_STYLE} />
            {!isEmpty(data) &&
                <Tooltip content={({payload, active}) => {
                    if (!active || !payload) {
                        return null;
                    }
                    
                    return (
                        <div className="trend-chart-tooltip">
                            <div className="tooltip-time">{formatDateBy(payload[0].payload.time, timeFormat)}</div>
                            <div className="tooltip-counts">
                                {
                                    
                                    Object.values(SEVERITY_ITEMS).map(({value, color}) => {
                                        const {value: count} = payload.find(payloadItem => payloadItem.dataKey === value);
                                        
                                        return (
                                            <LegendItem key={value} color={color} title={count} isDarkMode />
                                        )
                                    })
                                }
                            </div>
                        </div>
                    )
                }} />
            }
        </AreaChart>
    </ResponsiveContainer>
);

export default TrendsChart;