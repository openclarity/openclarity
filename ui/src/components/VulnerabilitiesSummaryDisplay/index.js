import React from 'react';
import classnames from 'classnames';
import Tooltip from 'components/Tooltip';
import Icon, { ICON_NAMES } from 'components/Icon';
import { SEVERITY_ITEMS, CIS_SEVERITY_ITEMS } from 'utils/systemConsts';

import './vulnerabilities-summary-display.scss';

export const CisBenchmarkLevelsDisplay = ({id, levels, withTotal}) => (
    <VulnerabilitiesSummaryDisplay
        id={id}
        vulnerabilities={levels || []}
        icon={ICON_NAMES.ALERT}
        severityItems={CIS_SEVERITY_ITEMS}
        severityKey="level"
        withTotal={withTotal}
    />
)

const TotalDisplay = ({id, vulnerabilities}) => {
    const totalTooltipId = `vulnerability-summery-total-${id}`;
    const totalCount = vulnerabilities.reduce((acc, {count=0}) => acc + count, 0);

    return (
        <React.Fragment>
            <div className="vulnerability-summery-total" data-tip data-for={totalTooltipId}>{`Total: ${totalCount}`}</div>
            <Tooltip id={totalTooltipId} text={`Total: ${totalCount}`} placement="left" />
        </React.Fragment>
    )
}

const VulnerabilitiesSummaryDisplay = ({id, vulnerabilities, withTotal, isNarrow=false, icon=ICON_NAMES.BUG, severityItems=SEVERITY_ITEMS, severityKey="severity"}) => (
    <div className={classnames("vulnerabilities-summary-display", {narrow: isNarrow})}>
        {withTotal && <TotalDisplay id={id} vulnerabilities={vulnerabilities} />}
        {
            Object.values(severityItems).map(({value, label, color}) => {
                const {count=0} = vulnerabilities.find(item => item[severityKey] === value) || {};
                const tooltipId = `vulnerability-summery-${id}-${value}`;

                return (
                    <React.Fragment key={value}>
                        <div data-tip data-for={tooltipId} className="vulnerabilities-summary-item">
                            <Icon name={icon} className={classnames("vulnerability-icon", {"zero-count": count === 0})} style={{color}} />
                            <div className="vulnerability-count">{count}</div>
                        </div>
                        <Tooltip id={tooltipId} text={`${label}: ${count}`} />
                    </React.Fragment>
                )
            })
        }
    </div>
)

export default VulnerabilitiesSummaryDisplay;