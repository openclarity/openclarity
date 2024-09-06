import React from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import DetailsPageWrapper from "components/DetailsPageWrapper";
import { useQuery } from "@tanstack/react-query";
import ConfigurationActionsDisplay from "../ConfigurationActionsDisplay";
import QUERY_KEYS from "../../../../api/constants";
import { openClarityApi } from "../../../../api/openClarityApi";
import TabConfiguration from "./TabConfiguration";
import "./configuration-details.scss";

const DetailsContent = ({ data }) => {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const params = useParams();

  const id = params["id"];
  const innerTab = params["*"];

  return (
    <div className="configuration-details-content">
      <div className="configuration-details-content-header">
        <ConfigurationActionsDisplay
          data={data}
          onDelete={() => navigate(pathname.replace(`/${id}/${innerTab}`, ""))}
        />
      </div>
      <TabConfiguration data={data} />
    </div>
  );
};

const ConfigurationDetails = () => {
  const { id } = useParams();
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.scanConfigs, id],
    queryFn: () =>
      openClarityApi.getScanConfigsScanConfigID(
        id,
        "id,name,scanTemplate/scope,scanTemplate/assetScanTemplate/scanFamiliesConfig,scheduled,scanTemplate/maxParallelScanners,scanTemplate/assetScanTemplate/scannerInstanceCreationConfig",
      ),
  });

  return (
    <DetailsPageWrapper
      className="configuration-details-page-wrapper"
      backTitle="Scan configurations"
      getTitleData={(data) => ({ title: data?.name })}
      detailsContent={DetailsContent}
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default ConfigurationDetails;
