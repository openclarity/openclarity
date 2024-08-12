import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import ConfigurationsTable from './ConfigurationsTable';
import ConfigurationDetails from './ConfigurationDetails';

const Configurations = () => (
    <ListAndDetailsRouter listComponent={ConfigurationsTable} detailsComponent={ConfigurationDetails} />
)


export default Configurations;