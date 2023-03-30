import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';

const SecretsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Fingerprint",
            id: "fingerprint",
            accessor: "findingInfo.fingerprint",
            disableSort: true,
            width: 200
        },
        {
            Header: "Description",
            id: "description",
            accessor: "findingInfo.description",
            disableSort: true
        },
        {
            Header: "File path",
            id: "findingInfo",
            accessor: "findingInfo.filePath",
            disableSort: true,
            width: 200
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            tableTitle="secrets"
            findingsObjectType="Secret"
        />
    )
}

export default SecretsTable;