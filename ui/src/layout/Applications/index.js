import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import InnerAppLink from 'components/InnerAppLink';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import ApplicationsTable from './ApplicationsTable';
import ApplicationDetails from './ApplicationDetails';

export const ApplicationsLink = ({count, applicationResourceID, packageID, title}) => {
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        setFilters(filtersDispatch, {type: FILTER_TYPES.APPLICATIONS, filters: {applicationResourceID, packageID, title}, isSystem: true});
    }

    return (
        <InnerAppLink pathname={ROUTES.APPLICATIONS} onClick={onClick}>{count}</InnerAppLink>
    )
}

const Applications = () => (
    <ListAndDetailsRouter listComponent={ApplicationsTable} detailsComponent={ApplicationDetails} />
)

export default Applications;