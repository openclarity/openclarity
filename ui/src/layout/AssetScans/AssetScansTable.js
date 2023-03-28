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
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], []);

    return (
        <TablePage
            columns={columns}
            url={APIS.ASSET_SCANS}
            expand="scan,target"
            select="id,target,summary,scan"
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSET_SCANS}
            withMargin
        />
    )
}

export default AssetScansTable;