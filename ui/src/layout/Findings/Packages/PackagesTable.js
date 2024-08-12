import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import { OPERATORS } from 'components/Filter';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import FindingsTablePage from '../FindingsTablePage';

const PackagesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Package name",
            id: "name",
            sortIds: ["findingInfo.name"],
            accessor: "findingInfo.name"
        },
        {
            Header: "Version",
            id: "version",
            sortIds: ["findingInfo.version"],
            accessor: "findingInfo.version"
        },
        {
            Header: "Language",
            id: "language",
            sortIds: ["findingInfo.language"],
            accessor: "findingInfo.language"
        },
        {
            Header: "Licenses",
            id: "licenses",
            sortIds: ["findingInfo.licenses"],
            Cell: ({row}) => {
                const {licenses} = row.original.findingInfo || {};

                return (
                    <ExpandableList items={licenses || []} />
                )
            }
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            filterType={FILTER_TYPES.FINDINGS_PACKAGES}
            filtersConfig={[
                {value: "findingInfo.name", label: "Package name", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.version", label: "Version", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.language", label: "Language", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.licenses", label: "License", operators: [
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]}
            ]}
            tableTitle="packages"
            findingsObjectType="Package"
        />
    )
}

export default PackagesTable;
