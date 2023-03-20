import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import MisconfigurationTable from './MisconfigurationTable';
import MisconfigurationDetails from './MisconfigurationDetails';

const Misconfigurations = () => (
    <ListAndDetailsRouter listComponent={MisconfigurationTable} detailsComponent={MisconfigurationDetails} />
)


export default Misconfigurations;