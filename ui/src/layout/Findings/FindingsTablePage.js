import React from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { FILTER_TYPES } from 'context/FiltersProvider';

const FindingsTablePage = ({tableTitle, findingsObjectType, columns}) => (
    <TablePage
        columns={columns}
        url={APIS.FINDINGS}
        tableTitle={tableTitle}
        filterType={FILTER_TYPES.FINDINGS}
        filters={`findingInfo/objectType eq '${findingsObjectType}'`}
        expand="asset,scan"
    />
)

export default FindingsTablePage;