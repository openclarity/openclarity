import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import RootkitsTable from './RootkitsTable';
import RootkitDetails from './RootkitDetails';

const Rootkits = () => (
    <ListAndDetailsRouter listComponent={RootkitsTable} detailsComponent={RootkitDetails} />
)


export default Rootkits;