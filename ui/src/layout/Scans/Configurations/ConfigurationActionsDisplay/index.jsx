import React, { useState } from "react";
import { isNull } from "lodash";
import { ICON_NAMES } from "components/Icon";
import IconWithTooltip from "components/IconWithTooltip";
import Modal from "components/Modal";
import { BoldText } from "utils/utils";
import {
  useModalDisplayDispatch,
  MODAL_DISPLAY_ACTIONS,
} from "layout/Scans/ScanConfigWizardModal/ModalDisplayProvider";
import { useMutation } from "@tanstack/react-query";
import { openClarityApi } from "../../../../api/openClarityApi";
import "./configuration-actions-display.scss";

const ConfigurationActionsDisplay = ({ data, onDelete }) => {
  const modalDisplayDispatch = useModalDisplayDispatch();
  const setScanConfigFormData = (data) =>
    modalDisplayDispatch({
      type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA,
      payload: data,
    });

  const { id, scheduled } = data;
  const { operationTime, cronLine } = scheduled || {};

  const disableStartScan =
    Date.now() - new Date(operationTime).valueOf() <= 0 && !cronLine;

  const [deleteConfirmationData, setDeleteConfirmationData] = useState(null);
  const closeDeleteConfirmation = () => setDeleteConfirmationData(null);

  const scanConfigDeleteMutation = useMutation({
    mutationFn: () => openClarityApi.deleteScanConfigsScanConfigID(id),
    onSettled: () => {
      onDelete();
    },
  });

  const scanConfigPatchMutation = useMutation({
    mutationFn: () =>
      openClarityApi.patchScanConfigsScanConfigID(id, {
        scheduled: {
          ...scheduled,
          operationTime: new Date().toISOString(),
        },
        disabled: false,
      }),
  });

  return (
    <>
      <div className="configuration-actions-display">
        <IconWithTooltip
          tooltipId={`${id}-start`}
          tooltipText={
            disableStartScan
              ? "A scan cannot be started for non-repetitive scan configs that are scheduled for a future time. You can edit or duplicate it to scan 'Now'."
              : "Start scan"
          }
          name={ICON_NAMES.PLAY}
          onClick={(event) => {
            event.stopPropagation();
            event.preventDefault();

            scanConfigPatchMutation.mutate();
          }}
          disabled={disableStartScan}
        />
        <IconWithTooltip
          tooltipId={`${id}-duplicate`}
          tooltipText="Duplicate scan configuration"
          name={ICON_NAMES.DUPLICATE}
          onClick={(event) => {
            event.stopPropagation();
            event.preventDefault();

            setScanConfigFormData({ ...data, id: null, name: "" });
          }}
        />
        <IconWithTooltip
          tooltipId={`${id}-edit`}
          tooltipText="Edit scan configuration"
          name={ICON_NAMES.EDIT}
          onClick={(event) => {
            event.stopPropagation();
            event.preventDefault();

            setScanConfigFormData(data);
          }}
        />
        <IconWithTooltip
          tooltipId={`${id}-delete`}
          tooltipText="Delete scan configuration"
          name={ICON_NAMES.DELETE}
          onClick={(event) => {
            event.stopPropagation();
            event.preventDefault();

            setDeleteConfirmationData(data);
          }}
        />
      </div>
      {!isNull(deleteConfirmationData) && (
        <Modal
          title="Delete confirmation"
          isMediumTitle
          className="scan-config-delete-confirmation"
          onClose={closeDeleteConfirmation}
          height={250}
          doneTitle="Delete"
          onDone={() => {
            scanConfigDeleteMutation.mutate();

            closeDeleteConfirmation();
          }}
        >
          <span>{`Once `}</span>
          <BoldText>{deleteConfirmationData.name}</BoldText>
          <span>{` will be deleted, the action cannot be reverted`}</span>
          <br />
          <span>{`Are you sure you want to delete ${deleteConfirmationData.name}?`}</span>
        </Modal>
      )}
    </>
  );
};

export default ConfigurationActionsDisplay;
