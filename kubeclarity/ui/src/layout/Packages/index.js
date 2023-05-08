import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import InnerAppLink from 'components/InnerAppLink';
import { ROUTES } from 'utils/systemConsts';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import PackagesTable from './PackagesTable';
import PackageDetails from './PackageDetails';

export const PackagesLink = ({value, applicationID, applicationResourceID, title}) => {
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        setFilters(filtersDispatch, {type: FILTER_TYPES.PACKAGES, filters: {applicationID, applicationResourceID, title}, isSystem: true});
    }

    return (
        <InnerAppLink pathname={ROUTES.PACKAGES} onClick={onClick}>{value}</InnerAppLink>
    )
}
    
const Packages = () => (
    <ListAndDetailsRouter listComponent={PackagesTable} detailsComponent={PackageDetails} />
)

export default Packages;