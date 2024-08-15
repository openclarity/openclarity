import React from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { TitleValueDisplayColumn } from "components/TitleValueDisplay";
import DoublePaneDisplay from "components/DoublePaneDisplay";
import Button from "components/Button";
import Title from "components/Title";
import Loader from "components/Loader";
import ConfigurationReadOnlyDisplay from "layout/Scans/ConfigurationReadOnlyDisplay";
import { ROUTES, APIS } from "utils/systemConsts";
import { formatNumber } from "utils/utils";
import {
  useFilterDispatch,
  setFilters,
  FILTER_TYPES,
} from "context/FiltersProvider";
import { useQuery } from "@tanstack/react-query";
import { openClarityApi } from "../../../../api/openClarityApi";

function ConfigurationScansDisplay({ configId, configName }) {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const filtersDispatch = useFilterDispatch();

  const scansFilter = `scanConfig/id eq '${configId}'`;

  const onScansClick = () => {
    setFilters(filtersDispatch, {
      type: FILTER_TYPES.SCANS,
      filters: {
        filter: scansFilter,
        name: configName,
        suffix: "configuration",
        backPath: pathname,
      },
      isSystem: true,
    });

    navigate(ROUTES.SCANS);
  };

  const { data, isError, isLoading } = useQuery({
    queryKey: [APIS.SCANS, scansFilter],
    queryFn: () => openClarityApi.getScans(scansFilter, "count", true, 1),
  });

  if (isError) {
    return null;
  }

  if (isLoading) {
    return <Loader absolute={false} small />;
  }

  return (
    <>
      <Title medium>Configuration's scans</Title>
      <Button
        onClick={onScansClick}
      >{`See all scans (${formatNumber(data.data.count || 0)})`}</Button>
    </>
  );
}

const TabConfiguration = ({ data }) => {
  const { id, name } = data || {};

  return (
    <DoublePaneDisplay
      leftPaneDisplay={() => (
        <>
          <Title medium>Configuration</Title>
          <TitleValueDisplayColumn>
            <ConfigurationReadOnlyDisplay configData={data} />
          </TitleValueDisplayColumn>
        </>
      )}
      rightPlaneDisplay={() => (
        <ConfigurationScansDisplay configId={id} configName={name} />
      )}
    />
  );
};

export default TabConfiguration;
