import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "assets";

const LOCATION_SORT_IDS = ["targetInfo.location"];

const AssetsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "instanceID",
            sortIds: ["targetInfo.instanceID"],
            accessor: "targetInfo.instanceID"
        },
        {
            Header: "Type",
            id: "objectType",
            sortIds: ["targetInfo.objectType"],
            accessor: "targetInfo.objectType"
        },
        {
            Header: "Location",
            id: "location",
            sortIds: LOCATION_SORT_IDS,
            accessor: "targetInfo.location"
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
            defaultSortBy={{sortIds: LOCATION_SORT_IDS, desc: false}}
            withMargin
        />
    )
}

export default AssetsTable;