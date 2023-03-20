import React from 'react';
import ListAndDetailsRouter from 'components/ListAndDetailsRouter';
import SecretsTable from './SecretsTable';
import SecretDetails from './SecretDetails';

const Secrets = () => (
    <ListAndDetailsRouter listComponent={SecretsTable} detailsComponent={SecretDetails} />
)


export default Secrets;