import React from "react";
import { useLocation, useParams } from "react-router-dom";
import DetailsPageWrapper from "components/DetailsPageWrapper";
import TabbedPage from "components/TabbedPage";
import { formatDate } from "utils/utils";
import {
  ScanDetails as ScanDetailsTab,
  Findings,
} from "layout/detail-displays";
import { useQuery } from "@tanstack/react-query";
import QUERY_KEYS from "../../../api/constants";
import { openClarityApi } from "../../../api/openClarityApi";
import ScanActionsDisplay from "./ScanActionsDisplay";

export const SCAN_DETAILS_PATHS = {
  FINDINGS: "findings",
};

const DetailsContent = ({ data, refetch }) => {
  const { pathname } = useLocation();

  const { id, name } = data;

  return (
    <TabbedPage
      basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
      items={[
        {
          id: "general",
          title: "Scan details",
          isIndex: true,
          component: () => (
            <ScanDetailsTab scanData={data} withAssetScansLink />
          ),
        },
        {
          id: "findings",
          title: "Findings",
          path: SCAN_DETAILS_PATHS.FINDINGS,
          component: () => (
            <Findings
              findingsSummary={data?.summary}
              findingsFilter={`foundBy/scan/id eq '${id}'`}
              findingsFilterTitle={name}
            />
          ),
        },
      ]}
      headerCustomDisplay={() => (
        <ScanActionsDisplay data={data} refetch={refetch} />
      )}
      withInnerPadding={false}
    />
  );
};

const ScanDetails = () => {
  const { id } = useParams();
  const { data, isError, isLoading, refetch } = useQuery({
    queryKey: [QUERY_KEYS.scans, id],
    queryFn: () =>
      openClarityApi.getScansScanID(
        id,
        "id,name,scanConfig,scope,assetScanTemplate,maxParallelScanners,startTime,endTime,summary,status",
      ),
  });

  return (
    <DetailsPageWrapper
      className="scan-details-page-wrapper"
      backTitle="Scans"
      getTitleData={({ name, startTime }) => ({
        title: name,
        subTitle: formatDate(startTime),
      })}
      detailsContent={(props) => (
        <DetailsContent {...props} refetch={() => refetch()} />
      )}
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default ScanDetails;
