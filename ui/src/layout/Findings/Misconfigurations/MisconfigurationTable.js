import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';
import { MISCONFIGURATION_SEVERITY_MAP } from './utils';

const MisconfigurationsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Test ID",
            id: "testId",
            sortIds: ["findingInfo.testID"],
            accessor: "findingInfo.testID"
        },
        {
            Header: "Severity",
            id: "severity",
            sortIds: ["findingInfo.severity"],
            accessor: original => MISCONFIGURATION_SEVERITY_MAP[original.findingInfo?.severity]
        },
        {
            Header: "Description",
            id: "description",
            sortIds: ["findingInfo.testDescription"],
            accessor: "findingInfo.testDescription",
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