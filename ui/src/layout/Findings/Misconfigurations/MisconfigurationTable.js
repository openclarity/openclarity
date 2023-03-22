import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';

const MisconfigurationsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "File path",
            id: "path",
            accessor: "findingInfo.path",
            disableSort: true
        },
        {
            Header: "Description",
            id: "description",
            accessor: "findingInfo.description",
            disableSort: true
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            tableTitle="misconfigurations"
            findingsObjectType="Misconfiguration"
        />
    )
}

export default MisconfigurationsTable;