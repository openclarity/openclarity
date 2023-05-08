import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import InnerAppLink from 'components/InnerAppLink';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import ApplicationResourcesTable from './ApplicationResourcesTable';
import ApplicationResourcesDetails from './ApplicationResourcesDetails';

export const ApplicationResourcesLink = ({count, applicationID, packageID, title}) => {
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        setFilters(filtersDispatch, {type: FILTER_TYPES.APPLICATION_RESOURCES, filters: {applicationID, packageID, title}, isSystem: true});
    }

    return (
        <InnerAppLink pathname={ROUTES.APPLICATION_RESOURCES} onClick={onClick}>{count}</InnerAppLink>
    )
}

const ApplicationResources = () => (
    <ListAndDetailsRouter listComponent={ApplicationResourcesTable} detailsComponent={ApplicationResourcesDetails} />
)

export default ApplicationResources;