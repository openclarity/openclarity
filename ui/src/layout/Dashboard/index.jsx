import React from "react";
import Loader from "components/Loader";
import { useQuery } from "@tanstack/react-query";
import { openClarityApi } from "../../api/openClarityApi";
import QUERY_KEYS from "../../api/constants";
import {
  AssetCounterDisplay,
  FindingCounterDisplay,
  ScanCounterDisplay,
} from "./CounterDisplay";
import FindingsTrendsWidget from "./FindingsTrendsWidget";
import RiskiestRegionsWidget from "./RiskiestRegionsWidget";
import RiskiestAssetsWidget from "./RiskiestAssetsWidget";
import FindingsImpactWidget from "./FindingsImpactWidget";
import EmptyScansDisplay from "./EmptyScansDisplay";
import "./dashboard.scss";

function Dashboard() {
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.scans],
    queryFn: () => openClarityApi.getScans(undefined, "count", true, 1),
  });

  if (isLoading) {
    return <Loader />;
  }

  if (isError) {
    return null;
  }

  if (data.data.count === 0) {
    return <EmptyScansDisplay />;
  }

  return (
    <div className="dashboard-page-wrapper">
      <ScanCounterDisplay />
      <AssetCounterDisplay />
      <FindingCounterDisplay />
      <RiskiestRegionsWidget className="riskiest-regions" />
      <FindingsTrendsWidget className="findings-trend" />
      <RiskiestAssetsWidget className="riskiest-assets" />
      <FindingsImpactWidget className="findings-impact" />
    </div>
  );
}

export default Dashboard;
