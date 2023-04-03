import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';

const SecretsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Fingerprint",
            id: "fingerprint",
            sortIds: ["findingInfo.fingerprint"],
            accessor: "findingInfo.fingerprint",
            width: 200
        },
        {
            Header: "Description",
            id: "description",
            sortIds: ["findingInfo.description"],
            accessor: "findingInfo.description"
        },
        {
            Header: "File path",
            id: "findingInfo",
            sortIds: ["findingInfo.filePath"],
            accessor: "findingInfo.filePath",
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