import React from 'react';
import { useNavigate } from 'react-router-dom';
import Icon, { ICON_NAMES } from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';
import { OPERATORS } from 'components/Filter';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS, ROUTES } from 'utils/systemConsts';
import WidgetWrapper from '../WidgetWrapper';

import './fixable-vulnerabilities-widget.scss';

const HeaderItemDisplay = ({icon, countTitle, count, onClick}) => (
    <div className="fixable-vulneraboilities-widget-header-item-display" onClick={onClick}>
        <Icon name={icon} />
        <div className="counter-wrapper">
            <div className="counter-count">{count}</div>
            <div className="counter-title">{countTitle}</div>
        </div>
    </div>
)

const Bar = ({count, total, title, color, onClick}) => {
    const percent = !total ? 0 : count/total*100;
    const rgbColor = `${parseInt(color.slice(1, 3), 16)}, ${parseInt(color.slice(3, 5), 16)}, ${parseInt(color.slice(5, 7), 16)}`

    return (
        <div className="vulnerabilities-bar-wrapper" onClick={onClick}>
            <Icon name={ICON_NAMES.BUG} style={{color}} />
            <div className="counters-container">
                <TooltipWrapper className="counters-wrapper" tooltipId={`dashboard-fixample-bar-tooltip-${title}`} tooltipText={`${count}/${total}`} >
                    <div className="count-counter">{count}</div>
                    <div className="total-counter">{`/${total}`}</div>
                </TooltipWrapper>
            </div>
            <div className="bar-container" style={{backgroundColor: `rgba(${rgbColor}, 0.3)`}}>
                <div className="bar-filler" style={{width: `${percent}%`, backgroundColor: `rgba(${rgbColor}, 1)`}}></div>  
            </div>
            <div className="bar-title">{title}</div>
        </div>
    )
}

const getTotalCountByKey = (data, key) => (data || []).reduce((acc, curr) => acc + (curr[key] || 0), 0);

const FixableVulnerabilitiesWidget = ({data}) => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const onItemClick = ({hasFix, severity}) => {
        const filtersData = [];
        if (hasFix || !!severity) {
            filtersData.push({scope: "hasFixVersion", operator: OPERATORS.is.value, value: [true]});

            if (!!severity) {
                filtersData.push({scope: "vulnerabilitySeverity", operator: OPERATORS.is.value, value: [severity]});
            }
        }

        setFilters(filtersDispatch, {type: FILTER_TYPES.VULNERABILITIES, filters: filtersData});
        navigate(ROUTES.VULNERABILITIES);
    }

    return (
        <div>
            <div className="widget-header">
                <HeaderItemDisplay
                    icon={ICON_NAMES.BANDAID}
                    countTitle="Fix available"
                    count={getTotalCountByKey(data, "countWithFix")}
                    onClick={() => onItemClick({hasFix: true})}
                />
                <HeaderItemDisplay
                    icon={ICON_NAMES.VULNERABILITY}
                    countTitle="Vulnerabilities"
                    count={getTotalCountByKey(data, "countTotal")}
                    onClick={() => onItemClick({hasFix: false})}
                />
            </div>
            <div className="widget-content">
                {
                    Object.values(SEVERITY_ITEMS).map(({value, label, color}) => {
                        const {countTotal, countWithFix} = (data || []).find(({severity}) => severity === value) || {};
    
                        return (
                            <Bar
                                key={value}
                                title={label}
                                color={color}
                                total={countTotal || 0}
                                count={countWithFix || 0}
                                onClick={() => onItemClick({hasFix: false, severity: value})}
                            />
                        )
                    })
                }
            </div>
        </div>
    )
}

const FixableVulnerabilitiesWidgetWrapper = ({refreshTimestamp}) => (
    <WidgetWrapper
        className="fixable-vulneraboilities-widget"
        title="Fixable vulnerabilities"
        url="dashboard/vulnerabilitiesWithFix"
        widget={FixableVulnerabilitiesWidget}
        refreshTimestamp={refreshTimestamp}
    />
)

export default FixableVulnerabilitiesWidgetWrapper;

