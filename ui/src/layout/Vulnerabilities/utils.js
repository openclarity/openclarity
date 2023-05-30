import React from 'react';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
import InnerAppLink from 'components/InnerAppLink';
import { ROUTES, SEVERITY_ITEMS } from 'utils/systemConsts';
import { BoldText } from 'utils/utils';
import { ApplicationsLink as GeneralApplicationsLink } from 'layout/Applications';
import { ApplicationResourcesLink as GeneralApplicationResourcesLink } from 'layout/ApplicationResources';

export const PackagesLink = ({packageVersion, packageName}) => {
    const filtersDispatch = useFilterDispatch();

    const filterData = [
        {scope: "packageVersion", operator: OPERATORS.is.value, value: [packageVersion]},
        {scope: "packageName", operator: OPERATORS.is.value, value: [packageName]}
    ];

    const onClick = () => {
        setFilters(filtersDispatch, {type: FILTER_TYPES.PACKAGES, filters: filterData, isSystem: false});
    }

    return (
        <InnerAppLink pathname={ROUTES.PACKAGES} onClick={onClick}>{packageVersion}</InnerAppLink>
    )
}

const getTitle = name => `vulnerability: ${name}`;

export const ApplicationsLink = ({packageID, applications, vulnerabilityName}) => (
    <GeneralApplicationsLink count={applications} packageID={packageID} title={getTitle(vulnerabilityName)} />
)

export const ApplicationResourcesLink = ({packageID, applicationResources, vulnerabilityName}) => (
    <GeneralApplicationResourcesLink count={applicationResources} packageID={packageID} title={getTitle(vulnerabilityName)} />
)

export const CvssScoreMessage = ({cvssScore, cvssSeverity}) => {
    const {label} = SEVERITY_ITEMS[cvssSeverity] || {};

    return (
        <span>
            {`Although the CVSS base impact score is `}
            <BoldText>{cvssScore}</BoldText>
            {` (${label || cvssSeverity}), the linux distribution severity reflects the risk more accurately.`}
        </span>
    )
}