import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { getScanName, getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "asset scans";

const AssetScansTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Asset name",
            id: "name",
            accessor: "target.targetInfo.instanceID",
            disableSort: true
        },
        {
            Header: "Asset type",
            id: "type",
            accessor: "target.targetInfo.objectType",
            disableSort: true
        },
        {
            Header: "Asset location",
            id: "location",
            accessor: "target.targetInfo.location",
            disableSort: true
        },
        {
            Header: "Scan",
            id: "scan",
            accessor: original => {
                const {startTime, scanConfigSnapshot} = original.scan;
                
                return getScanName({name: scanConfigSnapshot?.name, startTime});
            },
            disableSort: true
        },
        getVulnerabilitiesColumnConfigItem({tableTitle: TABLE_TITLE, idKey: "scan.id", summaryKey: "scan.summary"}),
        ...getFindingsColumnsConfigList({tableTitle: TABLE_TITLE, summaryKey: "scan.summary"})
    ], []);

    return (
        <TablePage
            columns={columns}
            url={APIS.ASSET_SCANS}
            expand="scan,target"
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSET_SCANS}
            withMargin
        />
    )
}

export default AssetScansTable;