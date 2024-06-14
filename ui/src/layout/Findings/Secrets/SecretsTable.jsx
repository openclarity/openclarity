import React, { useMemo } from 'react';
import { getScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
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
        ...getScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            filterType={FILTER_TYPES.FINDINGS_SECRETS}
            filtersConfig={[
                {value: "findingInfo.fingerprint", label: "Fingerprint", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.description", label: "Description", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.filePath", label: "File path", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]}
            ]}
            tableTitle="secrets"
            findingsObjectType="Secret"
        />
    )
}

export default SecretsTable;