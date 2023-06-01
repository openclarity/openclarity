import React, { useMemo } from 'react';
import TablePage from 'components/TablePage';
import ExpandableList from 'components/ExpandableList';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, getAssetColumnsFiltersConfig,
    findingsColumnsFiltersConfig, vulnerabilitiesCountersColumnsFiltersConfig, formatTagsToStringsList } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "assets";

const LOCATION_SORT_IDS = ["targetInfo.location"];

const AssetsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "instanceID",
            sortIds: ["targetInfo.instanceID"],
            accessor: "targetInfo.instanceID"
        },
        {
            Header: "Labels",
            id: "tags",
            sortIds: ["targetInfo.tags"],
            Cell: ({row}) => {
                const {tags} = row.original.targetInfo;
                
                return (
                    <ExpandableList items={formatTagsToStringsList(tags)} withTagWrap />
                )
            },
            alignToTop: true
        },
        {
            Header: "Type",
            id: "objectType",
            sortIds: ["targetInfo.objectType"],
            accessor: "targetInfo.objectType"
        },
        {
            Header: "Location",
            id: "location",
            sortIds: LOCATION_SORT_IDS,
            accessor: "targetInfo.location"
        },
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], []);
    
    return (
        <TablePage
            columns={columns}
            url={APIS.ASSETS}
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
