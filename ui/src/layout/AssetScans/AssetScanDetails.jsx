import React from "react";
import { useLocation, useParams } from "react-router-dom";
import DetailsPageWrapper from "components/DetailsPageWrapper";
import TabbedPage from "components/TabbedPage";
import { formatDate } from "utils/utils";
import { Findings } from "layout/detail-displays";
import { useQuery } from "@tanstack/react-query";
import QUERY_KEYS from "../../api/constants";
import { openClarityApi } from "../../api/openClarityApi";
import TabAssetScanDetails from "./TabAssetScanDetails";

const ASSET_SCAN_DETAILS_PATHS = {
  FINDINGS: "findings",
};

const DetailsContent = ({ data }) => {
  const { pathname } = useLocation();

  const { id, summary } = data;

  return (
    <TabbedPage
      basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
      items={[
        {
          id: "general",
          title: "Asset scan details",
          isIndex: true,
          component: () => <TabAssetScanDetails data={data} />,
        },
        {
          id: "findings",
          title: "Findings",
          path: ASSET_SCAN_DETAILS_PATHS.FINDINGS,
          component: () => (
            <Findings
              findingsSummary={summary}
              findingsFilter={`foundBy/id eq '${id}'`}
              findingsFilterTitle={`AssetScan ${id}`} // TODO(sambetts) replace with name maybe
            />
          ),
        },
      ]}
      withInnerPadding={false}
    />
  );
};

const AssetScanDetails = () => {
  const { id } = useParams();
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.assetScans, id],
    queryFn: () =>
      openClarityApi.getAssetScansAssetScanID(
        id,
        "id,scan,asset,summary,status,stats,sbom/status,vulnerabilities/status,exploits/status,misconfigurations/status,secrets/status,malware/status,rootkits/status",
        "scan($select=id,name,startTime,endTime),asset($select=id,assetInfo),status,stats",
      ),
  });

  return (
    <DetailsPageWrapper
      backTitle="Asset scans"
      getTitleData={({ scan, asset }) => {
        const { startTime, name } = scan || {};

        return {
          title: asset?.assetInfo?.instanceID,
          subTitle: `scanned by '${name}' on ${formatDate(startTime)}`,
        };
      }}
      detailsContent={(props) => <DetailsContent {...props} />}
      withPadding
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default AssetScanDetails;
