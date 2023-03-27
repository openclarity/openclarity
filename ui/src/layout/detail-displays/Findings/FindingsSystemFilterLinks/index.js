import React from 'react';
import { useNavigate } from 'react-router-dom';
import Button from 'components/Button';
import Icon from 'components/Icon';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';

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

const FindingsSystemFilterLinks = ({totalVulnerabilitiesCount, findingsSummary}) => {
    const {appRoute: vulnerabilitiesAppRoute, title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

    return (
        <div className="findings-system-filters-links">
            <div className="findings-system-filters-title">See other findings:</div>
            <FindingFilterLink
                icon={vulnerabilitiesIcon}
                title={`${totalVulnerabilitiesCount} ${vulnerabilitiesTitle}`}
                appRoute={vulnerabilitiesAppRoute}
            />
            {
                Object.keys(FINDINGS_MAPPING).map(findingType => {
                    const {totalKey, title, icon, appRoute} = FINDINGS_MAPPING[findingType];
    
                    return (
                        <FindingFilterLink key={findingType} icon={icon} title={`${findingsSummary[totalKey] || 0} ${title}`} appRoute={appRoute} />
                    )
                })
            }
        </div>
    )
}

export default FindingsSystemFilterLinks;