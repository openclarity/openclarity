import React from 'react';
import classnames from 'classnames';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, ResponsiveContainer, Tooltip } from 'recharts';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import Icon from 'components/Icon';
import { APIS, FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { BoldText } from 'utils/utils';
import WidgetWrapper from '../WidgetWrapper';
import { LegendIcon } from '../utils';

import COLORS from 'utils/scss_variables.module.scss';

import './reskiest-regions-widget.scss';

const BAR_STACK_ID = 1;
const WIDGET_FINDINGS_TYPES= Object.keys(FINDINGS_MAPPING).filter(findingType => findingType !== FINDINGS_MAPPING.PACKAGES.value);

const TooltipCountItem = ({color, icon, value, title}) => (
    <div className="riskiest-regions-widget-tooltip-content-item">
        <Icon name={icon} size={18} style={{color}} />
        <div className="riskiest-regions-widget-tooltip-content-item-text">{`${value} ${title}`}</div>
    </div>
)

const CustomTooltip = ({active, payload}) => {
    if (active && payload && payload.length) {
        const {regionName, findingsCount} = payload[0].payload;
        const vulnerabilitiesCount = findingsCount.vulnerabilities || 0;
        const total = Object.values(findingsCount).reduce((acc, curr) => {
            return acc + curr;
        }, 0)

        const {darkColor: vulnerabilitiesColor, title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

        return (
            <div className="riskiest-regions-widget-tooltip">
                <div className="riskiest-regions-widget-tooltip-header">
                    <BoldText>{regionName}</BoldText>
                    <div style={{marginTop: "3px"}}>{`Total findings: `}<BoldText>{total}</BoldText></div>
                </div>
                <div className="riskiest-regions-widget-tooltip-content">
                    {vulnerabilitiesCount > 0 &&
                        <TooltipCountItem
                            title={vulnerabilitiesTitle}
                            icon={vulnerabilitiesIcon}
                            value={vulnerabilitiesCount}
                            color={vulnerabilitiesColor}
                        />
                    }
                    {
                        WIDGET_FINDINGS_TYPES.map(findingType => {
                            const {dataKey, title, icon, color, darkColor} = FINDINGS_MAPPING[findingType];
                            const value = findingsCount[dataKey] || 0;

                            if (value === 0) {
                                return null;
                            }
                            
                            return (
                                <TooltipCountItem key={findingType} title={title} icon={icon} value={value} color={darkColor || color} />
                            )
                        })
                    }
                </div>
            </div>
        )
    }

    return null;
}

const WidgetLegendIcon = ({title, icon, color}) => (
    <LegendIcon widgetName="riskiest-regions" title={title} icon={icon} color={color} />
)

const WidgetContent = ({data}) => {
    const {dataKey: vulnerabilitiesKey, color: vulnerabilitiesColor, title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

    return (
        <div style={{display: "flex", flexDirection: "column", height: "100%"}}>
            <div style={{display: "flex", justifyContent: "space-between"}}>
                <WidgetLegendIcon title={vulnerabilitiesTitle} icon={vulnerabilitiesIcon} color={vulnerabilitiesColor} />
                {
                    WIDGET_FINDINGS_TYPES.map((findingType, index) => {
                        const {title, icon, color} = FINDINGS_MAPPING[findingType];
                        
                        return (
                            <WidgetLegendIcon key={index} title={title} icon={icon} color={color} />
                        )
                    })
                }
            </div>
            <ResponsiveContainer width="100%" height="100%">
                <BarChart data={data} layout="vertical" barSize={10} margin={{top: 12, right: 10, left: 20, bottom: 60}}>
                    <CartesianGrid horizontal={false} style={{stroke: COLORS["color-grey-lighter"]}}/>
                    <XAxis type="number" tick={{fill: COLORS["color-grey"]}} style={{fontSize: "12px"}} />
                    <YAxis type="category" dataKey="regionName" tick={{fill: COLORS["color-grey-black"]}} style={{fontSize: "12px"}} />
                    <Tooltip
                        content={<CustomTooltip />}
                        wrapperStyle={{backgroundColor: "rgba(34, 37, 41, 0.95)", outline: "none", padding: "10px", color: "white", fontSize: "12px"}}
                        cursor={{fill: COLORS["color-grey-lighter"]}}
                    />
                    <Bar dataKey={`findingsCount.${vulnerabilitiesKey}`} stackId={BAR_STACK_ID} fill={vulnerabilitiesColor} />
                    {
                        WIDGET_FINDINGS_TYPES.map(findingType => {
                            const {dataKey, color} = FINDINGS_MAPPING[findingType];
                            
                            return (
                                <Bar key={findingType} dataKey={`findingsCount.${dataKey}`} stackId={BAR_STACK_ID} fill={color} />
                            )
                        })
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