import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import AssetDetails from './AssetDetails';
import AssetsTable from './AssetsTable';

const Assets = () => (
    <ListAndDetailsRouter listComponent={AssetsTable} detailsComponent={AssetDetails} />
)

export default Assets;