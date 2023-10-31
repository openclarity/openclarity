import React from 'react';
import { TooltipWrapper } from 'components/Tooltip';
import Icon from 'components/Icon';
import { SEVERITY_ITEMS } from 'components/SeverityDisplay';
import { VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { formatNumber } from 'utils/utils';

import COLORS from 'utils/scss_variables.module.scss';

import './vulnerabilities-display.scss';

export const getTotlalVulnerabilitiesFromCounters = (counters={}) => (
    Object.values(VULNERABILITY_SEVERITY_ITEMS).reduce((acc, curr) => {
        return acc + (counters[curr.totalKey] || 0);
    }, 0)
)

export const VULNERABILITY_SEVERITY_ITEMS = {
    totalCriticalVulnerabilities: {
        totalKey: "totalCriticalVulnerabilities",
        countKey: "criticalVulnerabilitiesCount",
        title: "Critical",
        color: SEVERITY_ITEMS.CRITICAL.color,
        innerTextColor: "white"
    },
    totalHighVulnerabilities: {
        totalKey: "totalHighVulnerabilities",
        countKey: "highVulnerabilitiesCount",
        title: "High",
        color: SEVERITY_ITEMS.HIGH.color,
        innerTextColor: "white"
    },
    totalMediumVulnerabilities: {
        totalKey: "totalMediumVulnerabilities",
        countKey: "mediumVulnerabilitiesCount",
        title: "Medium",
        color: SEVERITY_ITEMS.MEDIUM.color,
        innerTextColor: COLORS["color-grey-black"]
    },
    totalLowVulnerabilities: {
        totalKey: "totalLowVulnerabilities",
        countKey: "lowVulnerabilitiesCount",
        title: "Low",
        color: SEVERITY_ITEMS.LOW.color,
        innerTextColor: COLORS["color-grey-black"]
    },
    totalNegligibleVulnerabilities: {
        totalKey: "totalNegligibleVulnerabilities",
        countKey: "negligibleVulnerabilitiesCount",
        title: "Negligible",
        color: SEVERITY_ITEMS.NEGLIGIBLE.color,
        innerTextColor: COLORS["color-grey-black"],
        backgroundColor: "transparent"
    }
}

const TooltipContentDisplay = ({total, counters}) => (
    <div className="vulnerabilities-minimized-tooltip-content">
        <div>{`Vulnerabilities: ${formatNumber(total || 0)}`}</div>
        <div className="vulnerabilities-tooltip-counters">
            {
                Object.values(VULNERABILITY_SEVERITY_ITEMS).map(({totalKey, color}) => (
                    <div key={totalKey} className="vulnerabilities-tooltip-counters-item">
                        <Icon name={VULNERABIITY_FINDINGS_ITEM.icon} style={{color}} size={18} /><span>{formatNumber(counters[totalKey] || 0)}</span>
                    </div>
                ))
            }
        </div>
    </div>
)

const MinimizedVulnerabilitiesDisplay = ({id, highestSeverity, total, counters}) => {
    const {color, innerTextColor, backgroundColor} = VULNERABILITY_SEVERITY_ITEMS[highestSeverity];

    return (
        <div className="vulnerabilities-minimized-display-wrapper">
            <TooltipWrapper tooltipId={`vulnerability-minimized-display-${id}`} tooltipText={<TooltipContentDisplay total={total} counters={counters} />}>
                <div className="vulnerabilities-minimized-display-summary-item" style={{color: innerTextColor, backgroundColor: backgroundColor || color}}>
                    {formatNumber(counters[highestSeverity] || 0)}
                </div>
            </TooltipWrapper>
        </div>
    )
}

const CounterItemDisplay = ({count, title, color}) => (
    <div className="vulnerabilities-display-counter-item">
        <div className="vulnerabilities-counter-item-count" style={{color}}>{count || 0}</div>
        <div className="vulnerabilities-counter-item-title">{title}</div>
    </div>
)

const VulnerabilitiesDisplay = ({highestSeverity, total, counters}) => {
    const {color} = VULNERABILITY_SEVERITY_ITEMS[highestSeverity];
    const {color: vulnerabilitiesColor, title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

    return (
        <div className="vulnerabilities-display-wrapper">
            <div className="vulnerabilities-display-summary">
                <Icon name={vulnerabilitiesIcon} style={{color}} size={30} />
                <CounterItemDisplay count={formatNumber(total)} title={vulnerabilitiesTitle} color={vulnerabilitiesColor} />
            </div>
            <div className="vulnerabilities-display-counters">
                {
                    Object.values(VULNERABILITY_SEVERITY_ITEMS).map(({totalKey, title, color}) => (
                        <CounterItemDisplay key={totalKey} count={formatNumber(counters[totalKey])} title={title} color={color} />
                    ))
                }
            </div>
        </div>
    )
}


const VulnerabilitiesDisplayWrapper = ({counters={}, isMinimized=false, minimizedTooltipId=null}) => {
    const total = getTotlalVulnerabilitiesFromCounters(counters);
    
    const highestSeverity = (Object.values(VULNERABILITY_SEVERITY_ITEMS).find(({totalKey}) =>
        !!counters[totalKey] && counters[totalKey] > 0) || VULNERABILITY_SEVERITY_ITEMS.totalNegligibleVulnerabilities).totalKey;

    const DisplayComponent = isMinimized ? MinimizedVulnerabilitiesDisplay : VulnerabilitiesDisplay;
    
    return (
        <DisplayComponent id={minimizedTooltipId} highestSeverity={highestSeverity} total={total} counters={counters} />
    )
}

export default VulnerabilitiesDisplayWrapper;
