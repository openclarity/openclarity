import React, { useState } from 'react';
import classnames from 'classnames';
import { orderBy } from 'lodash';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, ResponsiveContainer, Tooltip } from 'recharts';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import { APIS, FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { BoldText, formatNumber } from 'utils/utils';
import WidgetWrapper from '../WidgetWrapper';
import ChartTooltip from '../ChartTooltip';
import FindingsFilters from '../FindingsFilters';

import COLORS from 'utils/scss_variables.module.scss';

import './reskiest-regions-widget.scss';

const BAR_STACK_ID = 1;
const WIDGET_FINDINGS_ITEMS = [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING).filter(({value}) => value !== FINDINGS_MAPPING.PACKAGES.value)];

const TooltipHeader = ({data}) => {
    const {regionName, ...countData} = data;
    
    const total = Object.values(countData).reduce((acc, curr) => {
        return acc + curr;
    }, 0)

    return (
        <>
            <BoldText>{regionName}</BoldText>
            <div style={{marginTop: "3px"}}>{`Total findings: `}<BoldText>{formatNumber(total)}</BoldText></div>
        </>
    )
}

const WidgetContent = ({data}) => {
    const [selectedFilters, setSelectedFilters] = useState([
        ...WIDGET_FINDINGS_ITEMS.map(({dataKey}) => dataKey)
    ]);

    const formattedData = orderBy(data || [], ({findingsCount}) => {
        return Object.keys(findingsCount || {}).reduce((acc, currFindingKey) => {
            const count = findingsCount[currFindingKey] || 0;
            
            return acc + (selectedFilters.includes(currFindingKey) ? count : 0);
        }, 0);
    }, ["desc"]).map(({regionName, findingsCount}) => ({regionName, ...findingsCount}));

    return (
        <div style={{display: "flex", flexDirection: "column", height: "100%"}}>
            <FindingsFilters
                widgetName="riskiest-regions"
                findingsItems={WIDGET_FINDINGS_ITEMS}
                findingKeyName="dataKey"
                selectedFilters={selectedFilters}
                setSelectedFilters={setSelectedFilters}
            />
            <ResponsiveContainer width="100%" height="100%">
                <BarChart data={formattedData} layout="vertical" barSize={10} margin={{top: 12, right: 10, left: 20, bottom: 60}}>
                    <CartesianGrid horizontal={false} style={{stroke: COLORS["color-grey-lighter"]}}/>
                    <XAxis type="number" tick={{fill: COLORS["color-grey"]}} style={{fontSize: "12px"}} />
                    <YAxis type="category" dataKey="regionName" tick={{fill: COLORS["color-grey-black"]}} style={{fontSize: "12px"}} />
                    <Tooltip
                        content={props => <ChartTooltip {...props} widgetFindings={WIDGET_FINDINGS_ITEMS} headerDisplay={TooltipHeader} countKeyName="dataKey" />}
                        wrapperStyle={{backgroundColor: "rgba(34, 37, 41, 0.95)", outline: "none", padding: "10px", color: "white", fontSize: "12px"}}
                        cursor={{fill: COLORS["color-grey-lighter"]}}
                    />
                    {
                        WIDGET_FINDINGS_ITEMS.map(({dataKey, color}) => (
                            selectedFilters.includes(dataKey) && <Bar key={dataKey} dataKey={dataKey} stackId={BAR_STACK_ID} fill={color} />
                        ))
                    }
                </BarChart>
            </ResponsiveContainer>
        </div>
    )
}

const RiskiestRegionsWidget = ({className}) => {
    const [{data, error, loading}] = useFetch(APIS.DASHBOARD_RISKIEST_REGIONS, {urlPrefix: "ui"});

    return (
        <WidgetWrapper className={classnames("riskiest-regions-widget", className)} title="Riskiest regions">
            {loading ? <Loader /> : (error ? null : <WidgetContent data={data?.regions} />)}
        </WidgetWrapper>
    )
}

export default RiskiestRegionsWidget;