import React from 'react';
import { useNavigate } from 'react-router-dom';
import classnames from 'classnames';
import { orderBy, isEmpty } from 'lodash';
import { PieChart, Pie, Cell } from 'recharts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
import { ROUTES } from 'utils/systemConsts';
import WidgetWrapper from '../WidgetWrapper';
import { NO_DATA } from '../utils';

import './packages-pie-widget.scss';

import COLORS from 'utils/scss_variables.module.scss';

const ITEMS_LIMIT = 4;

const PIE_COLORS = [
    COLORS["color-main"],
    COLORS["color-dash-blue"],
    COLORS["color-dash-blue-light"],
    COLORS["color-dash-blue-lighter"]
];

const formatDataForDisplay = (data, titleKey) => {
    const orderedData = orderBy((data || []), ["count"], ["desc"]).map(item => ({count: item.count, title: item[titleKey]}));

    if (orderedData.length <= ITEMS_LIMIT) {
        return orderedData;
    }

    const displayItems = orderedData.slice(0, ITEMS_LIMIT - 1);
    const othersItems = orderedData.slice(ITEMS_LIMIT - 1);
    const othersCount = othersItems.reduce((acc, {count=0}) => acc + count, 0);

    return [...displayItems, {title: "Others", count: othersCount, filterNamesNot: [displayItems.map(item => item.title)]}];
}

const LegendItem = ({title, count, color, onClick}) => (
    <div className="legend-item">
        <div className="legend-item-indecator" style={{backgroundColor: color}}></div>
        <div className="legend-item-data">
            <div className={classnames("legend-item-name", {clickable: !!onClick})} onClick={onClick}>{title}</div>
            <div className="legend-item-count">{count}</div>
        </div>
    </div>
);

const LegendLine = ({children}) => <div className="pie-legend-line">{children}</div>

const RADIAN = Math.PI / 180;
const PieLabel = ({cx, cy, midAngle, innerRadius, outerRadius, percent, index}) => {
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);
    
    return (
        <text x={x} y={y} className="pie-label" textAnchor="middle" dominantBaseline="central">
            {`${(percent * 100).toFixed(0)}%`}
        </text>
    );
};

const CART_SIZE = 146;
const INNER_RADIUS = 44;
const PackagesPieWidget = ({data, itemTitleKey, filterName}) => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const onLegendItemClick = ({name, filterNamesNot}) => {
        const isIsNotFilter = !isEmpty(filterNamesNot);
        
        setFilters(filtersDispatch, {type: FILTER_TYPES.PACKAGES, filters: [{
            scope: filterName,
            operator: isIsNotFilter ? OPERATORS.isNot.value : OPERATORS.is.value,
            value: isIsNotFilter ? filterNamesNot : [name]
        }]});

        navigate(ROUTES.PACKAGES)
    }

    const displayData = formatDataForDisplay(data, itemTitleKey);
    const noData = isEmpty(displayData);

    const emptyDisplayChartData = [{count: 1}];
    const emptyDisplayLegendDataItem = {count: 0, title: NO_DATA}
    const emptyDisplayLegendData = [emptyDisplayLegendDataItem, emptyDisplayLegendDataItem, emptyDisplayLegendDataItem, emptyDisplayLegendDataItem];
    
    return (
        <div>
            <div className="pie-content-wrapper">
                <div className="pie-content">
                    <div className="pie-total-display" style={{top: `${CART_SIZE / 2}px`, left: `${CART_SIZE / 2}px`, width: `${2 * INNER_RADIUS - 5}px`}}>
                        <div className="pie-total-title">Total</div>
                        <div className="pie-total-count">{noData ? NO_DATA : displayData.reduce((acc, {count=0}) => acc + count, 0)}</div>
                    </div>
                    <PieChart width={CART_SIZE} height={CART_SIZE}>
                        <Pie data={noData ? emptyDisplayChartData : displayData} innerRadius={INNER_RADIUS} outerRadius={CART_SIZE/2} dataKey="count" label={noData ? undefined : PieLabel} labelLine={false}>
                            {displayData.map((entry, index) => <Cell key={`cell-${index}`} fill={PIE_COLORS[index]} />)}
                            {noData && <Cell fill="white" stroke={COLORS["color-main"]} />}
                        </Pie>
                    </PieChart>
                </div>
            </div>
            <div className="pie-legend">
                <LegendLine>
                    {
                        (noData ? emptyDisplayLegendData : displayData).slice(0, 2).map(({title, count, filterNamesNot}, index) => (
                            <LegendItem
                                key={index}
                                title={title}
                                count={count}
                                color={PIE_COLORS[index]}
                                onClick={noData ? undefined : () => onLegendItemClick({name: title, filterNamesNot})}
                            />
                        ))
                    }
                </LegendLine>
                <LegendLine>
                    {
                        (noData ? emptyDisplayLegendData : displayData).slice(2, 4).map(({title, count, filterNamesNot}, index) => (
                            <LegendItem
                                key={index}
                                title={title}
                                count={count}
                                color={PIE_COLORS[index + 2]}
                                onClick={noData ? undefined : () => onLegendItemClick({name: title, filterNamesNot})}
                            />
                        ))
                    }
                </LegendLine>
            </div>
        </div>
    )
}

const PackagesPieWidgetWrapper = ({title, url, itemTitleKey, filterName, refreshTimestamp}) => (
    <WidgetWrapper
        className="packages-pie-widget"
        title={title}
        url={url}
        widget={props => <PackagesPieWidget {...props} itemTitleKey={itemTitleKey} filterName={filterName} />}
        refreshTimestamp={refreshTimestamp}
    />
)

export default PackagesPieWidgetWrapper;