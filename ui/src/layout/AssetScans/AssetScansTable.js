import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { getScanName, getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import StatusIndicator from './StatusIndicator';

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
        {
            Header: "Scan status",
            id: "status",
            accessor: original => {
                const {state, errors} = original?.status?.general || {};
                
                return <StatusIndicator state={state} errors={errors} tooltipId={original.id} />;
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
            select="id,target,summary,scan,status"
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSET_SCANS}
            withMargin
        />
    )
}

export default AssetScansTable;