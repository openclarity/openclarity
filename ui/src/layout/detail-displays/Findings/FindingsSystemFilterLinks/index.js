import React from 'react';
import { useNavigate } from 'react-router-dom';
import Button from 'components/Button';
import Icon from 'components/Icon';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { formatNumber } from 'utils/utils';
import { getFindingsAbsolutePathByFindingType } from 'layout/Findings';

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
        {
            [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING)].map(({value, totalKey, title, icon}) => {
                const LinkTitle = VULNERABIITY_FINDINGS_ITEM.value === value ? `${formatNumber(totalVulnerabilitiesCount)} ${title}` :
                    `${!!findingsSummary ? (formatNumber(findingsSummary[totalKey] || 0)) : 0} ${title}`;

                return (
                    <FindingFilterLink key={value} icon={icon} title={LinkTitle} appRoute={getFindingsAbsolutePathByFindingType(value)} />
                )
            })
        }
    </div>
)

export default FindingsSystemFilterLinks;