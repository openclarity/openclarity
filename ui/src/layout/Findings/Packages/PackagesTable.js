import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';

const PackagesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Package name",
            id: "name",
            accessor: "findingInfo.name",
            disableSort: true
        },
        {
            Header: "Version",
            id: "version",
            accessor: "findingInfo.version",
            disableSort: true
        },
        {
            Header: "Languege",
            id: "languege",
            accessor: "findingInfo.language",
            disableSort: true
        },
        {
            Header: "Licenses",
            id: "licenses",
            Cell: ({row}) => {
                const {licenses} = row.original.findingInfo || {};

                return (
                    <ExpandableList items={licenses || []} />
                )
            },
            disableSort: true
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