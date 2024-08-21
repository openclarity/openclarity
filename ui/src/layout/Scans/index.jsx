import React, { useCallback, useEffect } from "react";
import { useLocation, useSearchParams } from "react-router-dom";
import { isNull } from "lodash";
import { usePrevious } from "hooks";
import TabbedPage from "components/TabbedPage";
import Loader from "components/Loader";
import EmptyDisplay from "components/EmptyDisplay";
import { useQuery } from "@tanstack/react-query";
import QUERY_KEYS from "../../api/constants";
import { openClarityApi } from "../../api/openClarityApi";
import ScanConfigWizardModal from "./ScanConfigWizardModal";
import {
  ModalDisplayProvider,
  useModalDisplayState,
  useModalDisplayDispatch,
  MODAL_DISPLAY_ACTIONS,
} from "./ScanConfigWizardModal/ModalDisplayProvider";
import Scans from "./Scans";
import Configurations from "./Configurations";
import { SCANS_PATHS } from "./utils";

export { SCANS_PATHS };

export const OPEN_CONFIG_FORM_PARAM = "openConfigForm";

const ScansTabbedPage = () => {
  const { pathname } = useLocation();
  const [searchParams] = useSearchParams();
  const openConfigForm = searchParams.get(OPEN_CONFIG_FORM_PARAM);
  const prevOpenConfigForm = usePrevious(openConfigForm);

  const { modalDisplayData } = useModalDisplayState();
  const modalDisplayDispatch = useModalDisplayDispatch();
  const closeDisplayModal = () =>
    modalDisplayDispatch({ type: MODAL_DISPLAY_ACTIONS.CLOSE_DISPLAY_MODAL });
  const openDisplayModal = useCallback(
    () =>
      modalDisplayDispatch({
        type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA,
        payload: {},
      }),
    [modalDisplayDispatch],
  );

  useEffect(() => {
    if (!prevOpenConfigForm && openConfigForm) {
      openDisplayModal();
    }
  }, [prevOpenConfigForm, openConfigForm, openDisplayModal]);

  const {
    data: scansData,
    isError: isScansError,
    isLoading: isScansLoading,
    refetch: refetchScans,
    isRefetching: isScansRefetching,
  } = useQuery({
    queryKey: [QUERY_KEYS.scans],
    queryFn: () => openClarityApi.getScans(undefined, "count", true, 1),
  });

  const {
    data: scanConfigsData,
    isError: isScanConfigsError,
    isLoading: isScanConfigsLoading,
    refetch: refetchScanConfigs,
    isRefetching: isScanConfigsRefetching,
  } = useQuery({
    queryKey: [QUERY_KEYS.scanConfigs],
    queryFn: () => openClarityApi.getScanConfigs(undefined, "count", true, 1),
  });

  if (
    isScansLoading ||
    isScanConfigsLoading ||
    isScansRefetching ||
    isScanConfigsRefetching
  ) {
    return <Loader />;
  }

  if (isScansError || isScanConfigsError) {
    return null;
  }

  return (
    <>
      {scansData.data.count === 0 && scanConfigsData.data.count === 0 ? (
        <EmptyDisplay
          message={
            <>
              <div>No scans detected.</div>
              <div>Create your first scan configuration to see issues.</div>
            </>
          }
          title="New scan configuration"
          onClick={openDisplayModal}
        />
      ) : (
        <TabbedPage
          redirectTo={`${pathname}/${SCANS_PATHS.SCANS}`}
          items={[
            {
              id: "scans",
              title: "Scans",
              path: SCANS_PATHS.SCANS,
              component: Scans,
            },
            {
              id: "configs",
              title: "Configurations",
              path: SCANS_PATHS.CONFIGURATIONS,
              component: Configurations,
            },
          ]}
          withStickyTabs
        />
      )}
      {!isNull(modalDisplayData) && (
        <ScanConfigWizardModal
          initialData={modalDisplayData}
          onClose={closeDisplayModal}
          onSubmitSuccess={() => {
            closeDisplayModal();
            refetchScans();
            refetchScanConfigs();
          }}
        />
      )}
    </>
  );
};

const ScansWrapper = () => (
  <ModalDisplayProvider>
    <ScansTabbedPage />
  </ModalDisplayProvider>
);

export default ScansWrapper;
