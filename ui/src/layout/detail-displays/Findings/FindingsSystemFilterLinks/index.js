import React from 'react';
import { useNavigate } from 'react-router-dom';
import Button from 'components/Button';
import Icon from 'components/Icon';
import { ROUTES, FINDINGS_MAPPING, VULNERABILITIES_ICON_NAME } from 'utils/systemConsts';
import { FINDINGS_PATHS } from 'layout/Findings'

import './findings-system-filter-links.scss';

const FindingFilterLink = ({icon, title, appRoute}) => {
    const navigate = useNavigate();

    return (
        <div className="finging-filter-link">
            <Icon name={icon} size={20} />
            <Button tertiary onClick={() => navigate(appRoute)}>{title}</Button>
        </div>
    )
}

const FindingsSystemFilterLinks = ({totalVulnerabilitiesCount, findingsSummary}) => (
    <div className="findings-system-filters-links">
        <div className="findings-system-filters-title">See other findings:</div>
        <FindingFilterLink
            icon={VULNERABILITIES_ICON_NAME}
            title={`${totalVulnerabilitiesCount} Vulnerabilities`}
            appRoute={`${ROUTES.FINDINGS}/${FINDINGS_PATHS.VULNERABILITIES}`}
        />
        {
            Object.keys(FINDINGS_MAPPING).map(findingType => {
                const {dataKey, title, icon, appRoute} = FINDINGS_MAPPING[findingType];

                return (
                    <FindingFilterLink key={findingType} icon={icon} title={`${findingsSummary[dataKey] || 0} ${title}`} appRoute={appRoute} />
                )
            })
        }
    </div>
)

export default FindingsSystemFilterLinks;