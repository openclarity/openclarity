import React from "react";
import DetailsPageWrapper from "components/DetailsPageWrapper";
import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import QUERY_KEYS from "../../api/constants";
import { openClarityApi } from "../../api/openClarityApi";

const FindingsDetailsPage = ({
  backTitle,
  getTitleData,
  detailsContent: DetailsContent,
}) => {
  const { id } = useParams();
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.findings, id],
    queryFn: () =>
      openClarityApi.getFindingsFindingID(
        id,
        "id,findingInfo,firstSeen,lastSeen",
      ),
  });

  return (
    <DetailsPageWrapper
      backTitle={backTitle}
      getTitleData={getTitleData}
      detailsContent={(props) => <DetailsContent {...props} />}
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default FindingsDetailsPage;
