import React, { useMemo } from 'react';
import { getScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
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
        ...getScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            filterType={FILTER_TYPES.FINDINGS_ROOTKITS}
            filtersConfig={[
                {value: "findingInfo.rootkitName", label: "Rootkit name", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.message", label: "Message", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]}
            ]}
            tableTitle="rootkits"
            findingsObjectType="Rootkit"
        />
    )
}

export default RootkitsTable;
