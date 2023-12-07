import React, { useCallback, useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, ResponsiveContainer, Tooltip } from 'recharts';
import classnames from 'classnames';
import moment from 'moment';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import DropdownSelect from 'components/DropdownSelect';
import { APIS, FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { formatDate } from 'utils/utils';
import WidgetWrapper from '../WidgetWrapper';
import ChartTooltip from '../ChartTooltip';
import FindingsFilters from '../FindingsFilters';

import COLORS from 'utils/scss_variables.module.scss';

import './findings-trends-widget.scss';

const WIDGET_FINDINGS_ITEMS = [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING)];

const calcRange = (unit, value) => ({endTime: moment().toISOString(), startTime: moment().subtract(unit, value).toISOString()})

const TIME_RANGES = {
    HOUR: {
        value: "HOUR",
        label: "Last hour",
        calc: () => calcRange(1, 'hours')
    },
    DAY: {
        value: "DAY",
        label: "Last day",
        calc: () => calcRange(1, 'days')
    },
    WEEK: {
        value: "WEEK",
        label: "Last week",
        calc: () => calcRange(7, 'days')
    },
    MONTH: {
        value: "MONTH",
        label: "Last month",
        calc: () => calcRange(1, 'months')
    },
    YEAR: {
        value: "YEAR",
        label: "Last year",
        calc: () => calcRange(1, 'years')
    }
}

const TooltipHeader = ({data}) => (<div>{data.formattedTime}</div>)

const WidgetChart = ({data, selectedFilters}) => {
    const formattedData = data?.reduce((acc, curr) => {
        const {findingType, trends} = curr;

        trends.forEach(({time, count}) => {
            const accTimeIndex = acc.findIndex(({time: accTime}) => time === accTime);
            const formattedTime = formatDate(time);
            
            acc = accTimeIndex < 0 ? [...acc, {time, formattedTime, [findingType]: count}] :
                [
                    ...acc.slice(0, accTimeIndex),
                    {...acc[accTimeIndex], [findingType]: count},
                    ...acc.slice(accTimeIndex + 1)
                ];
        });

        return acc;
    }, []);
    
    return (
        <div className="findings-trends-widget-chart" style={{width: "100%", height: "100%"}}>
            <ResponsiveContainer width="100%" height="100%">
                <LineChart data={formattedData} margin={{top: 5, right: 0, left: 0, bottom: 60}}>
                    <CartesianGrid vertical={false} style={{stroke: COLORS["color-grey-lighter"]}}/>
                    <XAxis dataKey="formattedTime" tick={{fill: COLORS["color-grey"]}} style={{fontSize: "12px"}} />
                    <YAxis tick={{fill: COLORS["color-grey"]}} style={{fontSize: "12px"}} />
                    <Tooltip
                        content={props => <ChartTooltip {...props} widgetFindings={WIDGET_FINDINGS_ITEMS} headerDisplay={TooltipHeader} countKeyName="typeKey" />}
                        wrapperStyle={{backgroundColor: "rgba(34, 37, 41, 0.95)", outline: "none", padding: "10px", color: "white", fontSize: "12px"}}
                        cursor={{fill: COLORS["color-grey-lighter"]}}
                    />
                    {
                        WIDGET_FINDINGS_ITEMS.map(({color, typeKey}) => (
                            selectedFilters.includes(typeKey) &&
                                <Line key={typeKey} type="linear" dataKey={typeKey} stroke={color} dot={false} strokeWidth={2} />
                        ))
                    }
                </LineChart>
            </ResponsiveContainer>
        </div>
    )
}

const FindingsTrendsWidget = ({className}) => {
    const {value, label} = TIME_RANGES.WEEK;
    const [selectedRange, setSelectedRange] = useState({value, label});

    const [{data, error, loading}, fetchData] = useFetch(APIS.DASHBOARD_FINDINGS_TRENDS, {loadOnMount: false});
    const updateChartData = useCallback(({startTime, endTime}) => fetchData({urlPrefix: "ui", queryParams: {startTime, endTime}}), [fetchData]);

    useEffect(() => {
        const {startTime, endTime} = TIME_RANGES[selectedRange.value].calc();
        updateChartData({startTime, endTime});
    }, [selectedRange.value, updateChartData]);

    const [selectedFilters, setSelectedFilters] = useState([
        ...WIDGET_FINDINGS_ITEMS.map(({typeKey}) => typeKey)
    ]);
    
    return (
        <WidgetWrapper className={classnames("findings-trends-widget", className)} title="Findings trends">
            <div className="findings-trends-widget-header">
                <FindingsFilters
                    widgetName="findings-trends"
                    findingsItems={WIDGET_FINDINGS_ITEMS}
                    findingKeyName="typeKey"
                    selectedFilters={selectedFilters}
                    setSelectedFilters={setSelectedFilters}
                />
                <DropdownSelect
                    name="timeRangeSelect"
                    value={selectedRange}
                    items={Object.values(TIME_RANGES).map(({value, label}) => ({value, label}))}
                    onChange={selectedItem =>  setSelectedRange(selectedItem)}
                />
            </div>
            {loading ? <Loader /> : (error ? null : <WidgetChart data={data} selectedFilters={selectedFilters} />)}
        </WidgetWrapper>
    )
}

export default FindingsTrendsWidget;
