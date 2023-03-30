import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';
import { MISCONFIGURATION_SEVERITY_MAP } from './utils';

const MisconfigurationsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Test ID",
            id: "testId",
            accessor: "findingInfo.testID",
            disableSort: true
        },
        {
            Header: "Severity",
            id: "severity",
            accessor: original => MISCONFIGURATION_SEVERITY_MAP[original.findingInfo?.severity],
            disableSort: true
        },
        {
            Header: "Description",
            id: "description",
            accessor: "findingInfo.testDescription",
            disableSort: true,
            width: 200
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