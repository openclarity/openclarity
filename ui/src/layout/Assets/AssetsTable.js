import React, { useMemo, useEffect, useCallback } from 'react';
import { isUndefined } from 'lodash';
import TablePage from 'components/TablePage';
import ExpandableList from 'components/ExpandableList';
import ToggleButton from 'components/ToggleButton';
import Loader from 'components/Loader';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, getAssetColumnsFiltersConfig,
    findingsColumnsFiltersConfig, vulnerabilitiesCountersColumnsFiltersConfig, formatTagsToStringsList, formatDate} from 'utils/utils';
import { useFilterDispatch, useFilterState, setFilters, FILTER_TYPES } from 'context/FiltersProvider';

const TABLE_TITLE = "assets";

const LOCATION_SORT_IDS = ["assetInfo.location"];

const ASSETS_FILTER_TYPE = FILTER_TYPES.ASSETS;

const AssetsTable = () => {
    const filtersDispatch = useFilterDispatch();
    const filtersState = useFilterState();

    const {customFilters} = filtersState[ASSETS_FILTER_TYPE];
    const {hideTerminated} = customFilters;

    const setHideTerminated = useCallback(hideTerminated => setFilters(filtersDispatch, {
        type: ASSETS_FILTER_TYPE,
        filters: {hideTerminated},
        isCustom: true
    }), [filtersDispatch]);
    
    useEffect(() => {
        if (isUndefined(hideTerminated)) {
            setHideTerminated(true);
        }
    }, [hideTerminated, setHideTerminated]);

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
        {
            Header: "Last Seen",
            id: "lastSeen",
            sortIds: ["lastSeen"],
            accessor: original => formatDate(original.lastSeen)
        },
        ...(hideTerminated ? [] : [{
            Header: "Terminated On",
            id: "terminatedOn",
            sortIds: ["terminatedOn"],
            accessor: original => formatDate(original?.terminatedOn)
        }]),
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE)
    ], [hideTerminated]);
    
    if (isUndefined(hideTerminated)) {
        return <Loader />;
    }

    return (
        <div style={{position: "relative"}}>
            <div style={{position: "absolute", top: 0, right: "30px", zIndex: 1, display: "flex", alignItems: "center"}}>
                <ToggleButton title="Hide terminated" checked={hideTerminated} onChange={setHideTerminated}/>
            </div>
            <TablePage
                columns={columns}
                url={APIS.ASSETS}
                select="id,assetInfo,summary,lastSeen,terminatedOn"
                tableTitle={TABLE_TITLE}
                filterType={ASSETS_FILTER_TYPE}
                filtersConfig={[
                    ...getAssetColumnsFiltersConfig(),
                    ...vulnerabilitiesCountersColumnsFiltersConfig,
                    ...findingsColumnsFiltersConfig
                ]}
                filters={hideTerminated ? ["(terminatedOn eq null)"] : null}
                defaultSortBy={{sortIds: ["lastSeen", "terminatedOn"], desc: true}}
                withMargin
            />
        </div>
    )
}

export default AssetsTable;
