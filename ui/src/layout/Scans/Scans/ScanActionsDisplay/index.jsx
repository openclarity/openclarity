import React from "react";
import { ICON_NAMES } from "components/Icon";
import IconWithTooltip from "components/IconWithTooltip";
import { SCAN_STATES } from "components/ScanProgressBar";
import { useMutation } from "@tanstack/react-query";
import { openClarityApi } from "../../../../api/openClarityApi";
import "./scan-actions-display.scss";

const ScanActionsDisplay = ({ data, onUpdate }) => {
  const {
    id,
    status: { state },
  } = data;

  const { mutate: stopScanMutation, isPending } = useMutation({
    mutationFn: () =>
      openClarityApi.patchScansScanID(id, {
        status: {
          state: SCAN_STATES.Aborted.state,
          reason: "Cancellation",
          message: "Scan has been aborted",
          lastTransitionTime: new Date().toISOString(),
        },
      }),
    onSuccess: () => {
      if (onUpdate !== undefined) {
        onUpdate();
      }
    },
  });

  if (
    [
      SCAN_STATES.Done.state,
      SCAN_STATES.Failed.state,
      SCAN_STATES.Aborted.state,
    ].includes(state) ||
    isPending
  ) {
    return null;
  }

  return (
    <div className="scan-actions-display">
      <IconWithTooltip
        tooltipId={`${id}-stop`}
        tooltipText="Stop scan"
        name={ICON_NAMES.STOP}
        onClick={(event) => {
          event.stopPropagation();
          event.preventDefault();

          stopScanMutation();
        }}
      />
    </div>
  );
};

export default ScanActionsDisplay;
