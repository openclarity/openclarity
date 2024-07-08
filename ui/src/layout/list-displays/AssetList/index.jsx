import React, { useMemo, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { isUndefined } from 'lodash';
import TablePage from 'components/TablePage';
import ExpandableList from 'components/ExpandableList';
import ToggleButton from 'components/ToggleButton';
import Loader from 'components/Loader';
import { APIS } from 'utils/systemConsts';
import { getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, getAssetColumnsFiltersConfig,
    findingsColumnsFiltersConfig, vulnerabilitiesCountersColumnsFiltersConfig, formatTagsToStringsList, formatDate} from 'utils/utils';
import { useFilterDispatch, useFilterState, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { getAssetName } from 'utils/utils';

const TABLE_TITLE = "assets";

const NAME_SORT_IDS = ["asset.assetInfo.instanceID", "asset.assetInfo.podName", "asset.assetInfo.dirName", "asset.assetInfo.imageID", "asset.assetInfo.containerName"];
const LABEL_SORT_IDS = ["asset.assetInfo.tags", "asset.assetInfo.labels"];
const LOCATION_SORT_IDS = ["asset.assetInfo.location"];

const ASSETS_FILTER_TYPE = FILTER_TYPES.ASSETS;

const AssetList = (props) => {
    const {findingId} = props;

    const navigate = useNavigate();
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
            sortIds: NAME_SORT_IDS,
            accessor: (original) => getAssetName(original.asset.assetInfo),
        },
        {
            Header: "Labels",
            id: "tags",
            sortIds: LABEL_SORT_IDS,
            Cell: ({row}) => {
                const {tags, labels} = row.original.asset.assetInfo;

                return (
                    <ExpandableList items={formatTagsToStringsList(tags ?? labels)} withTagWrap />
                )
            },
            alignToTop: true
        },
        {
            Header: "Type",
            id: "objectType",
            sortIds: ["asset.assetInfo.objectType"],
            accessor: "asset.assetInfo.objectType"
        },
        {
            Header: "Location",
            id: "location",
            sortIds: LOCATION_SORT_IDS,
            accessor: (original) => original.asset.assetInfo.location || original.asset.assetInfo.repoDigests?.[0],
        },
        {
            Header: "Last Seen",
            id: "lastSeen",
            sortIds: ["asset.lastSeen"],
            accessor: original => formatDate(original.asset.lastSeen)
        },
        ...(hideTerminated ? [] : [{
            Header: "Terminated On",
            id: "terminatedOn",
            sortIds: ["asset.terminatedOn"],
            accessor: original => formatDate(original?.asset.terminatedOn)
        }]),
        getVulnerabilitiesColumnConfigItem({tableTitle: TABLE_TITLE, withAssetPrefix: true}),
        ...getFindingsColumnsConfigList({tableTitle: TABLE_TITLE, withAssetPrefix: true})
    ], [hideTerminated]);

    if (isUndefined(hideTerminated)) {
        return <Loader />;
    }

    let filtersList = [`(finding.id eq '${findingId}')`]
    if (hideTerminated) {
        filtersList.push("(asset.terminatedOn eq null)");
    }
    let select = "asset.id,asset.assetInfo,asset.lastSeen,asset.summary"
    if (!hideTerminated) {
        select += ",asset.terminatedOn"
    }
    const expand = "asset"

    return (
        <div style={{position: "relative"}}>
            <div style={{position: "absolute", top: 0, right: "30px", zIndex: 1, display: "flex", alignItems: "center"}}>
                <ToggleButton title="Hide terminated" checked={hideTerminated} onChange={setHideTerminated}/>
            </div>
            <TablePage
                columns={columns}
                url={APIS.ASSET_FINDINGS}
                expand={expand}
                select={select}
                filters={filtersList.length > 0 ? [filtersList.join(" and ")] : null}
                tableTitle={TABLE_TITLE}
                filterType={ASSETS_FILTER_TYPE}
                filtersConfig={[
                    ...getAssetColumnsFiltersConfig({prefix: "asset.assetInfo"}),
                    ...vulnerabilitiesCountersColumnsFiltersConfig,
                    ...findingsColumnsFiltersConfig
                ]}
                onLineClick={({asset}) => navigate(`/${APIS.ASSETS}/${asset.id}`)}
                defaultSortBy={{sortIds: ["asset.lastSeen", "asset.terminatedOn"], desc: true}}
                defaulPageSize={10}
                withMargin
            />
        </div>
    )
}

export default AssetList;
