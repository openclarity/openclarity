import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, formatDate } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import StatusIndicator from './StatusIndicator';

const TABLE_TITLE = "asset scans";

const SCAN_START_TIME_SORT_IDS = ["scan.startTime"];

const AssetScansTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Asset name",
            id: "name",
            sortIds: ["target.targetInfo.instanceID"],
            accessor: "target.targetInfo.instanceID"
        },
        {
            Header: "Asset type",
            id: "type",
            sortIds: ["target.targetInfo.objectType"],
            accessor: "target.targetInfo.objectType"
        },
        {
            Header: "Asset location",
            id: "location",
            sortIds: ["target.targetInfo.location"],
            accessor: "target.targetInfo.location"
        },
        {
            Header: "Scan name",
            id: "scanName",
            sortIds: ["scan.scanConfigSnapshot.name"],
            accessor: "scan.scanConfigSnapshot.name"
        },
        {
            Header: "Scan start",
            id: "startTime",
            sortIds: SCAN_START_TIME_SORT_IDS,
            accessor: original => formatDate(original.scan?.startTime)
        },
        {
            Header: "Scan status",
            id: "status",
            sortIds: [
                "status.general.state",
                "status.general.errors"
            ],
            accessor: original => {
                const {state, errors} = original?.status?.general || {};
                
                return <StatusIndicator state={state} errors={errors} tooltipId={original.id} />;
            }
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
            defaultSortBy={{sortIds: SCAN_START_TIME_SORT_IDS, desc: true}}
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSET_SCANS}
            withMargin
        />
    )
}

export default AssetScansTable;