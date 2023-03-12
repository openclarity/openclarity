import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import ConfigurationsTable from './ConfigurationsTable';
import ConfigurationDetails from './ConfigurationDetails';


export const SCAN_CONFIGS_PATH = "configs";

const Configurations = () => (
    <ListAndDetailsRouter listComponent={ConfigurationsTable} detailsComponent={ConfigurationDetails} />
)


export default Configurations;