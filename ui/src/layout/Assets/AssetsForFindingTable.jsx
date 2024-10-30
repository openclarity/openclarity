import React, { useMemo, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { isUndefined } from "lodash";
import ExpandableList from "components/ExpandableList";
import ToggleButton from "components/ToggleButton";
import ContentContainer from "components/ContentContainer";
import Table from "components/Table";
import Loader from "components/Loader";
import {
  getFindingsColumnsConfigList,
  getVulnerabilitiesColumnConfigItem,
  formatTagsToStringsList,
  formatDate,
  getAssetName,
} from "utils/utils";
import { APIS } from "utils/systemConsts";
import {
  useFilterDispatch,
  useFilterState,
  setFilters,
  FILTER_TYPES,
} from "context/FiltersProvider";

const TABLE_TITLE = "assets";

const NAME_SORT_IDS = [
  "asset.assetInfo.instanceID",
  "asset.assetInfo.podName",
  "asset.assetInfo.dirName",
  "asset.assetInfo.imageID",
  "asset.assetInfo.containerName",
];
const LABEL_SORT_IDS = ["asset.assetInfo.tags", "asset.assetInfo.labels"];
const LOCATION_SORT_IDS = ["asset.assetInfo.location"];

const ASSETS_FILTER_TYPE = FILTER_TYPES.ASSETS;
const FINDINGS_FILTER_TYPE = FILTER_TYPES.FINDINGS_GENERAL;

const AssetsForFindingTable = (props) => {
  const { findingId } = props;

  const navigate = useNavigate();
  const filtersDispatch = useFilterDispatch();
  const filtersState = useFilterState();

  const { customFilters: assetCustomFilters } =
    filtersState[ASSETS_FILTER_TYPE];
  const { hideTerminated } = assetCustomFilters;

  const setHideTerminated = useCallback(
    (hideTerminated) =>
      setFilters(filtersDispatch, {
        type: ASSETS_FILTER_TYPE,
        filters: { hideTerminated },
        isCustom: true,
      }),
    [filtersDispatch],
  );

  useEffect(() => {
    if (isUndefined(hideTerminated)) {
      setHideTerminated(true);
    }
  }, [hideTerminated, setHideTerminated]);

  const { customFilters: findingCustomFilters } =
    filtersState[FINDINGS_FILTER_TYPE];
  const { showFindingCounts } = findingCustomFilters;

  const setShowFindingCounts = useCallback(
    (showFindingCounts) =>
      setFilters(filtersDispatch, {
        type: FINDINGS_FILTER_TYPE,
        filters: { showFindingCounts },
        isCustom: true,
      }),
    [filtersDispatch],
  );

  useEffect(() => {
    if (isUndefined(showFindingCounts)) {
      setShowFindingCounts(false);
    }
  }, [showFindingCounts, setShowFindingCounts]);

  const columns = useMemo(
    () => [
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
        Cell: ({ row }) => {
          const { tags, labels } = row.original.asset.assetInfo;

          return (
            <ExpandableList
              items={formatTagsToStringsList(tags ?? labels)}
              withTagWrap
            />
          );
        },
        alignToTop: true,
      },
      {
        Header: "Type",
        id: "objectType",
        sortIds: ["asset.assetInfo.objectType"],
        accessor: "asset.assetInfo.objectType",
      },
      {
        Header: "Location",
        id: "location",
        sortIds: LOCATION_SORT_IDS,
        accessor: (original) =>
          original.asset.assetInfo.location ||
          original.asset.assetInfo.repoDigests?.[0] ||
          original.asset.assetInfo.image?.repoDigests?.[0],
      },
      {
        Header: "Last Seen",
        id: "lastSeen",
        sortIds: ["asset.lastSeen"],
        accessor: (original) => formatDate(original.asset.lastSeen),
      },
      ...(hideTerminated
        ? []
        : [
            {
              Header: "Terminated On",
              id: "terminatedOn",
              sortIds: ["asset.terminatedOn"],
              accessor: (original) => formatDate(original?.asset.terminatedOn),
            },
          ]),
      ...(!showFindingCounts
        ? []
        : [
            getVulnerabilitiesColumnConfigItem({
              tableTitle: TABLE_TITLE,
              withAssetPrefix: true,
            }),
            ...getFindingsColumnsConfigList({
              tableTitle: TABLE_TITLE,
              withAssetPrefix: true,
            }),
          ]),
    ],
    [hideTerminated, showFindingCounts],
  );

  if (isUndefined(hideTerminated) || isUndefined(showFindingCounts)) {
    return <Loader />;
  }

  const expand = "asset";
  let filtersList = [`(finding.id eq '${findingId}')`];
  if (hideTerminated) {
    filtersList.push("(asset.terminatedOn eq null)");
  }
  let select = "asset.id,asset.assetInfo,asset.lastSeen";
  if (!hideTerminated) {
    select += ",asset.terminatedOn";
  }
  if (showFindingCounts) {
    select += ",asset.summary";
  }

  return (
    <div style={{ marginTop: "20px" }}>
      <div style={{ float: "right", marginRight: "36px" }}>
        <ToggleButton
          title="Show finding counts"
          checked={showFindingCounts}
          onChange={setShowFindingCounts}
        />
      </div>
      <div style={{ float: "right", marginRight: "36px" }}>
        <ToggleButton
          title="Hide terminated"
          checked={hideTerminated}
          onChange={setHideTerminated}
        />
      </div>
      <div style={{ position: "relative" }}>
        <ContentContainer withMargin>
          <Table
            paginationItemsName={TABLE_TITLE.toLowerCase()}
            filters={{
              ...(!!expand ? { $expand: expand } : {}),
              ...(!!select ? { $select: select } : {}),
              ...(filtersList.length > 0
                ? { $filter: filtersList.join(" and ") }
                : {}),
            }}
            noResultsTitle={TABLE_TITLE}
            onLineClick={({ asset }) => navigate(`/${APIS.ASSETS}/${asset.id}`)}
            columns={columns}
            url={APIS.ASSET_FINDINGS}
            defaultSortBy={{
              sortIds: ["asset.lastSeen", "asset.terminatedOn"],
              desc: true,
            }}
            defaultPageSize={10}
          />
        </ContentContainer>
      </div>
    </div>
  );
};

export default AssetsForFindingTable;
