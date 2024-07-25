import React from "react";
import { useFetch } from "hooks";
import { isUndefined } from "lodash";
import TitleValueDisplay, {
  TitleValueDisplayRow,
} from "components/TitleValueDisplay";
import Loader from "components/Loader";
import { APIS } from "utils/systemConsts";
import { formatNumber } from "utils/utils";

const AssetCountDisplay = (findingId) => {
  const filter = `finding/id eq '${findingId}'`;
  const [{ loading, data, error }] = useFetch(APIS.ASSET_FINDINGS, {
    queryParams: {
      $expand: "asset",
      $filter: filter,
      $count: true,
      $select: "count,asset/terminatedOn",
    },
  });

  if (error) {
    return null;
  }

  if (loading) {
    return <Loader absolute={false} small />;
  }

  let notTerminated = 0;
  if (data && data.items) {
    data.items.forEach((item) => {
      if (isUndefined(item.asset.terminatedOn)) {
        notTerminated++;
      }
    });
  }

  return (
    <TitleValueDisplayRow>
      <TitleValueDisplay title="Asset count">
        {formatNumber(notTerminated)} ({formatNumber(data?.count || 0)})
      </TitleValueDisplay>
    </TitleValueDisplayRow>
  );
};

export default AssetCountDisplay;
