import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, formatDate, getAssetColumnsFiltersConfig,
    findingsColumnsFiltersConfig, vulnerabilitiesCountersColumnsFiltersConfig, scanColumnsFiltersConfig, getAssetName } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import StatusIndicator, { STATUS_MAPPING } from './StatusIndicator';

const TABLE_TITLE = "asset scans";

const NAME_SORT_IDS = [
    "asset.assetInfo.instanceID",
    "asset.assetInfo.podName",
    "asset.assetInfo.dirName",
    "asset.assetInfo.imageID",
    "asset.assetInfo.containerName"
];
const SCAN_START_TIME_SORT_IDS = ["scan.startTime"];

const FILTER_SCAN_STATUSES = Object.keys(STATUS_MAPPING).map(statusKey => (
    {value: statusKey, label: STATUS_MAPPING[statusKey]?.title}
))

const AssetScansTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Asset name",
            id: "name",
            sortIds: NAME_SORT_IDS,
            accessor: (assetScan) => getAssetName(assetScan.asset.assetInfo),
        },
        {
            Header: "Asset type",
            id: "type",
            sortIds: ["asset.assetInfo.objectType"],
            accessor: "asset.assetInfo.objectType"
        },
        {
            Header: "Asset location",
            id: "location",
            sortIds: ["asset.assetInfo.location"],
            accessor: (assetScan) => assetScan.asset.assetInfo.location || assetScan.asset.assetInfo.repoDigests?.[0],
        },
        {
            Header: "Scan name",
            id: "scanName",
            sortIds: ["scan.name"],
            accessor: "scan.name"
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
                "status.state",
                "status.message"
            ],
            accessor: original => {
                const {state, message} = original?.status || {};
                
                return <StatusIndicator state={state} errors={message === undefined ? [] : [message]} tooltipId={original.id} />;
            }
        },
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], []);

    return (
        <TablePage
            columns={columns}
            url={APIS.ASSET_SCANS}
            expand="scan($select=name,startTime),asset($select=assetInfo)"
            select="id,asset,summary,scan,status"
            defaultSortBy={{sortIds: SCAN_START_TIME_SORT_IDS, desc: true}}
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSET_SCANS}
            filtersConfig={[
                ...getAssetColumnsFiltersConfig({prefix: "asset.assetInfo", withLabels: false}),
                ...scanColumnsFiltersConfig,
                {value: "status.state", label: "Scan status", operators: [
                    {...OPERATORS.eq, valueItems: FILTER_SCAN_STATUSES},
                    {...OPERATORS.ne, valueItems: FILTER_SCAN_STATUSES}
                ]},
                ...vulnerabilitiesCountersColumnsFiltersConfig,
                ...findingsColumnsFiltersConfig
            ]}
            withMargin
        />
    )
}

export default AssetScansTable;
