import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';

const RootkitsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Rootkit name",
            id: "rootkitName",
            accessor: "findingInfo.rootkitName",
            disableSort: true
        },
        {
            Header: "File path",
            id: "path",
            accessor: "findingInfo.path",
            disableSort: true
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            tableTitle="rootkits"
            findingsObjectType="Rootkit"
        />
    )
}

export default RootkitsTable;