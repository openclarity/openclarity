import React, { useState, useEffect } from 'react';
import { useFetch } from 'hooks';
import TimeFilter, { TIME_SELECT_ITEMS, getTimeFormat } from 'components/TimeFilter';
import PageContainer from 'components/PageContainer';
import Loader from 'components/Loader';
import { SEVERITY_ITEMS } from 'utils/systemConsts';
import WidgetTitle from '../WidgetTitle';
import TrendsChart from './TrendsChart';
import LegendItem from './LegendItem';

import './vulnerabilities-trend-widget.scss';

const formatChartData = data => (
    (data || []).map(({time, numOfVuls}) => {
        const counts = numOfVuls.reduce((acc, {count, severity}) => {
            return {...acc, [severity]: count || 0};
        }, {});

        return {time, ...counts};
    })
)

const VulnerabilitiesTrendWidget = ({refreshTimestamp}) => {
    const defaultTimeRange = TIME_SELECT_ITEMS.DAY;
    const [timeFilter, setTimeFilter] = useState({selectedRange: defaultTimeRange.value, ...defaultTimeRange.calc()});
    const {selectedRange, startTime, endTime} = timeFilter;

    const [{loading, data}, fetchData] = useFetch("dashboard/trends/vulnerabilities", {loadOnMount: false});

    useEffect(() => {
        fetchData({queryParams: {startTime, endTime}});
    }, [startTime, endTime, fetchData, refreshTimestamp]);
    
    return (
        <PageContainer className="vulnerabilities-trend-widget">
            <div className="trend-widget-header">
                <WidgetTitle>New vulnerabilities trends</WidgetTitle>
                <TimeFilter
                    selectedRange={selectedRange}
                    startTime={startTime}
                    endTime={endTime}
                    onChange={({selectedRange, endTime, startTime}) => setTimeFilter({selectedRange, startTime, endTime})}
                />
            </div>
            <div className="trend-widget-legend">
                {
                    Object.values(SEVERITY_ITEMS).map(({value, label, color}) => (
                        <LegendItem key={value} color={color} title={label} />
                    ))
                }
            </div>
            {loading ? <Loader /> :
                <TrendsChart
                    data={formatChartData(data)}
                    timeFormat={getTimeFormat(startTime, endTime)}
                />
            }
        </PageContainer>
    )
}

export default VulnerabilitiesTrendWidget;