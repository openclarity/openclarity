import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import FindingsTablePage from '../FindingsTablePage';

const RootkitsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Rootkit name",
            id: "rootkitName",
            sortIds: ["findingInfo.rootkitName"],
            accessor: "findingInfo.rootkitName"
        },
        {
            Header: "Message",
            id: "message",
            sortIds: ["findingInfo.message"],
            accessor: "findingInfo.message"
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            filterType={FILTER_TYPES.FINDINGS_ROOTKITS}
            tableTitle="rootkits"
            findingsObjectType="Rootkit"
        />
    )
}

export default RootkitsTable;