import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import InnerAppLink from 'components/InnerAppLink';
import VulnerabilitiesSummaryDisplay from 'components/VulnerabilitiesSummaryDisplay';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import VulnerabilitiesTable from './VulnerabilitiesTable';
import VulnerabilityDetails from './VulnerabilityDetails';

import './vulnerabilities.scss'

export const setVulnerabilitiesSystemFilters = (filtersDispatch, {applicationID, applicationResourceID, packageID, title}) => (
    setFilters(filtersDispatch, {type: FILTER_TYPES.VULNERABILITIES, filters: {applicationID, applicationResourceID, packageID, title}, isSystem: true})
)

export const VulnerabilitiesLink = ({id, vulnerabilities, title, applicationID, applicationResourceID, packageID}) => {
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        setVulnerabilitiesSystemFilters(filtersDispatch, {applicationID, applicationResourceID, packageID, title});
    }
    
    return (
        <InnerAppLink pathname={ROUTES.VULNERABILITIES} onClick={onClick} className="vulnerabilities-inner-app-link">
            <VulnerabilitiesSummaryDisplay id={id} vulnerabilities={vulnerabilities} withTotal />
        </InnerAppLink>
    )
}

const Vulnerabilities = () => (
    <ListAndDetailsRouter
        listComponent={VulnerabilitiesTable}
        detailsComponent={VulnerabilityDetails}
        detailsPath=":vulnerabilityId/:packageId"
    />
)

export default Vulnerabilities;