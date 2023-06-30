import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import ExpandableList from 'components/ExpandableList';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, getAssetColumnsFiltersConfig,
    findingsColumnsFiltersConfig, vulnerabilitiesCountersColumnsFiltersConfig, formatTagsToStringsList } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "assets";

const LOCATION_SORT_IDS = ["assetInfo.location"];

const AssetsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "instanceID",
            sortIds: ["assetInfo.instanceID"],
            accessor: "assetInfo.instanceID"
        },
        {
            Header: "Labels",
            id: "tags",
            sortIds: ["assetInfo.tags"],
            Cell: ({row}) => {
                const {tags} = row.original.assetInfo;
                
                return (
                    <ExpandableList items={formatTagsToStringsList(tags)} withTagWrap />
                )
            },
            alignToTop: true
        },
        {
            Header: "Type",
            id: "objectType",
            sortIds: ["assetInfo.objectType"],
            accessor: "assetInfo.objectType"
        },
        {
            Header: "Location",
            id: "location",
            sortIds: LOCATION_SORT_IDS,
            accessor: "assetInfo.location"
        },
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], []);
    
    return (
        <TablePage
            columns={columns}
            url={APIS.ASSETS}
            select="id,assetInfo,summary"
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.ASSETS}
            filtersConfig={[
                ...getAssetColumnsFiltersConfig(),
                ...vulnerabilitiesCountersColumnsFiltersConfig,
                ...findingsColumnsFiltersConfig
            ]}
            defaultSortBy={{sortIds: LOCATION_SORT_IDS, desc: false}}
            withMargin
        />
    )
}

export default AssetsTable;
