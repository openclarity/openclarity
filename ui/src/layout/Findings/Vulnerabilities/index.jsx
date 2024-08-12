import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import VulnerabilitiesTable from './VulnerabilitiesTable';
import VulnerabilityDetails from './VulnerabilityDetails';

const Vulnerabilities = () => (
    <ListAndDetailsRouter listComponent={VulnerabilitiesTable} detailsComponent={VulnerabilityDetails} />
)


export default Vulnerabilities;
