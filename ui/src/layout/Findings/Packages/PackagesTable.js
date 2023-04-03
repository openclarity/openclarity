import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
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
            tableTitle="packages"
            findingsObjectType="Package"
        />
    )
}

export default PackagesTable;