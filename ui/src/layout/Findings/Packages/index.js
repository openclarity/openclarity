import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import PackagesTable from './PackagesTable';
import PackageDetails from './PackageDetails';

const Packages = () => (
    <ListAndDetailsRouter listComponent={PackagesTable} detailsComponent={PackageDetails} />
)


export default Packages;