import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import ScansTable from './ScansTable';
import ScanDetails from './ScanDetails';

export const SCAN_SCANS_PATH = "scans";

const Scans = () => (
    <ListAndDetailsRouter listComponent={ScansTable} detailsComponent={ScanDetails} />
)


export default Scans;