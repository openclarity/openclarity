import React from "react";
import { useLocation, useParams } from "react-router-dom";
import DetailsPageWrapper from "components/DetailsPageWrapper";
import TabbedPage from "components/TabbedPage";
import {
  AssetDetails as AssetDetailsTab,
  Findings,
} from "layout/detail-displays";
import { useQuery } from "@tanstack/react-query";
import QUERY_KEYS from "../../api/constants";
import { openClarityApi } from "../../api/openClarityApi";

const ASSET_DETAILS_PATHS = {
  FINDINGS: "findings",
};

const DetailsContent = ({ data }) => {
  const { pathname } = useLocation();

  const { id, assetInfo, summary } = data || {};

  return (
    <TabbedPage
      basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
      items={[
        {
          id: "general",
          title: "Asset details",
          isIndex: true,
          component: () => (
            <AssetDetailsTab assetData={data} withAssetScansLink />
          ),
        },
        {
          id: "findings",
          title: "Findings",
          path: ASSET_DETAILS_PATHS.FINDINGS,
          component: () => (
            <Findings
              findingsSummary={summary}
              findingsFilter={`asset/id eq '${id}'`}
              findingsFilterTitle={assetInfo.instanceID}
              findingsFilterSuffix="asset"
            />
          ),
        },
      ]}
      withInnerPadding={false}
    />
  );
};

const AssetDetails = () => {
  const { id } = useParams();
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.assets, id],
    queryFn: () =>
      openClarityApi.getAssetsAssetID(
        id,
        "id,assetInfo,summary,firstSeen,lastSeen,terminatedOn",
      ),
  });

  return (
    <DetailsPageWrapper
      backTitle="Assets"
      getTitleData={({ assetInfo }) => ({ title: assetInfo?.instanceID })}
      detailsContent={(props) => <DetailsContent {...props} />}
      withPadding
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default AssetDetails;
