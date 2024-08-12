import React, { useMemo } from 'react';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { OPERATORS } from 'components/Filter';
import FindingsTablePage from '../FindingsTablePage';
import { MISCONFIGURATION_SEVERITY_MAP } from './utils';

const FILTER_SEVERITY_ITEMS = Object.keys(MISCONFIGURATION_SEVERITY_MAP)
    .map(severityKey => ({value: severityKey, label: MISCONFIGURATION_SEVERITY_MAP[severityKey]}));

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
            filterType={FILTER_TYPES.FINDINGS_MISCONFIGURATIONS}
            filtersConfig={[
                {value: "findingInfo.testID", label: "Test ID", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.severity", label: "Severity", operators: [
                    {...OPERATORS.eq, valueItems: FILTER_SEVERITY_ITEMS},
                    {...OPERATORS.ne, valueItems: FILTER_SEVERITY_ITEMS}
                ]},
                {value: "findingInfo.testDescription", label: "Description", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]}
            ]}
            tableTitle="misconfigurations"
            findingsObjectType="Misconfiguration"
        />
    )
}

export default MisconfigurationsTable;
