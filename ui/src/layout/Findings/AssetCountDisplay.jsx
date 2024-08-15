import React from "react";
import { isUndefined } from "lodash";
import TitleValueDisplay, {
  TitleValueDisplayRow,
} from "components/TitleValueDisplay";
import Loader from "components/Loader";
import { formatNumber } from "utils/utils";
import { useQuery } from "@tanstack/react-query";
import { openClarityApi } from "../../api/openClarityApi";
import QUERY_KEYS from "../../api/constants";

const AssetCountDisplay = (findingId) => {
  const filter = `finding/id eq '${findingId}'`;

  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.assetFindings, filter],
    queryFn: () =>
      openClarityApi.getAssetFindings(
        filter,
        "count,asset/terminatedOn",
        true,
        undefined,
        undefined,
        "asset",
      ),
  });

  if (isError) {
    return null;
  }

  if (isLoading) {
    return <Loader absolute={false} small />;
  }

  let notTerminated = 0;
  if (data && data.data.items) {
    data.data.items.forEach((item) => {
      if (isUndefined(item.asset.terminatedOn)) {
        notTerminated++;
      }
    });
  }

  return (
    <TitleValueDisplayRow>
      <TitleValueDisplay title="Asset count">
        {formatNumber(notTerminated)} ({formatNumber(data.data.count || 0)})
      </TitleValueDisplay>
    </TitleValueDisplayRow>
  );
};

export default AssetCountDisplay;
