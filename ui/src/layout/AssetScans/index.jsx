import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import AssetScanDetails from './AssetScanDetails';
import AssetScansTable from './AssetScansTable';

const AssetScans = () => (
    <ListAndDetailsRouter listComponent={AssetScansTable} detailsComponent={AssetScanDetails} />
)


export default AssetScans;