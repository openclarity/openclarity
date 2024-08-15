import React from "react";
import { useNavigate, useLocation } from "react-router-dom";
import Title from "components/Title";
import Button from "components/Button";
import Loader from "components/Loader";
import { ROUTES, APIS } from "utils/systemConsts";
import { formatNumber } from "utils/utils";
import {
  useFilterDispatch,
  setFilters,
  FILTER_TYPES,
} from "context/FiltersProvider";
import { useQuery } from "@tanstack/react-query";
import { openClarityApi } from "../../../api/openClarityApi";

export function AssetScansDisplay({ assetName, assetId }) {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const filtersDispatch = useFilterDispatch();

  const filter = `asset/id eq '${assetId}'`;

  const onAssetScansClick = () => {
    setFilters(filtersDispatch, {
      type: FILTER_TYPES.ASSET_SCANS,
      filters: { filter, name: assetName, suffix: "asset", backPath: pathname },
      isSystem: true,
    });

    navigate(ROUTES.ASSET_SCANS);
  };

  const { data, isLoading, isError } = useQuery({
    queryKey: [APIS.ASSET_SCANS, filter],
    queryFn: () => openClarityApi.getAssetScans(filter, "count", true, 1),
  });

  if (isError) {
    return null;
  }

  if (isLoading) {
    return <Loader absolute={false} small />;
  }

  return (
    <>
      <Title medium>Asset scans</Title>
      <Button
        onClick={onAssetScansClick}
      >{`See all asset scans (${formatNumber(data.data.count || 0)})`}</Button>
    </>
  );
}
