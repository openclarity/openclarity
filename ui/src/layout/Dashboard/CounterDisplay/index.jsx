import React from "react";
import { formatNumber } from "utils/utils";
import { useQuery } from "@tanstack/react-query";
import COLORS from "../../../utils/scss_variables.module.scss";
import { openClarityApi } from "../../../api/openClarityApi";
import QUERY_KEYS from "../../../api/constants";
import "./counter-display.scss";

export function ScanCounterDisplay() {
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.scans],
    queryFn: () =>
      openClarityApi.getScans(
        "status/state eq 'Aborted' or status/state eq 'Failed' or status/state eq 'Done'",
        "count",
        true,
        1,
      ),
  });

  return (
    <div
      className="dashboard-counter"
      style={{ background: COLORS["color-gradient-green"] }}
    >
      {!isLoading && !isError && (
        <div className="dashboard-counter-content">
          <div className="dashboard-counter-count">
            {formatNumber(data.data.count)}
          </div>
          Completed scans
        </div>
      )}
    </div>
  );
}

export function AssetCounterDisplay() {
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.assets],
    queryFn: () => openClarityApi.getAssets(undefined, "count", true, 1),
  });

  return (
    <div
      className="dashboard-counter"
      style={{ background: COLORS["color-gradient-blue"] }}
    >
      {!isLoading && !isError && (
        <div className="dashboard-counter-content">
          <div className="dashboard-counter-count">
            {formatNumber(data.data.count)}
          </div>
          Assets
        </div>
      )}
    </div>
  );
}

export function FindingCounterDisplay() {
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.findings],
    queryFn: () => openClarityApi.getFindings(undefined, "count", true, 1),
  });

  return (
    <div
      className="dashboard-counter"
      style={{ background: COLORS["color-gradient-yellow"] }}
    >
      {!isLoading && !isError && (
        <div className="dashboard-counter-content">
          <div className="dashboard-counter-count">
            {formatNumber(data.data.count)}
          </div>
          Findings
        </div>
      )}
    </div>
  );
}
