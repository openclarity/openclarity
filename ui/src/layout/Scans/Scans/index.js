import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import ScansTable from './ScansTable';
import ScanDetails from './ScanDetails';

const Scans = () => (
    <ListAndDetailsRouter listComponent={ScansTable} detailsComponent={ScanDetails} />
)

export default Scans;