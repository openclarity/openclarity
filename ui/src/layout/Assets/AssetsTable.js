import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "assets";

const AssetsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "name",
            accessor: "targetInfo.instanceID",
            disableSort: true
        },
        {
            Header: "Type",
            id: "type",
            accessor: "targetInfo.objectType",
            disableSort: true
        },
        {
            Header: "Location",
            id: "location",
            accessor: "targetInfo.location",
            disableSort: true
        },
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], []);

    return (
        <TablePage
            columns={columns}
            url={APIS.ASSETS}
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSETS}
            withMargin
        />
    )
}

export default AssetsTable;