import React from "react";
import { useLocation } from "react-router-dom";
import TabbedPage from "components/TabbedPage";
import FindingsDetailsPage from "../FindingsDetailsPage";
import TabSecretDetails from "./TabSecretDetails";
import AssetsForFindingTable from "layout/Assets/AssetsForFindingTable";

const SECRET_DETAILS_PATHS = {
  ASSET_LIST: "assets",
};

const DetailsContent = ({ data }) => {
  const { pathname } = useLocation();

  const { id } = data;

  return (
    <TabbedPage
      basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
      items={[
        {
          id: "general",
          title: "Secret details",
          isIndex: true,
          component: () => <TabSecretDetails data={data} />,
        },
        {
          id: "assets",
          title: "Assets",
          path: SECRET_DETAILS_PATHS.ASSET_LIST,
          component: () => <AssetsForFindingTable findingId={id} />,
        },
      ]}
      withInnerPadding={false}
    />
  );
};

const SecretDetails = () => (
  <FindingsDetailsPage
    backTitle="Secrets"
    getTitleData={({ findingInfo }) => ({ title: findingInfo.fingerprint })}
    detailsContent={DetailsContent}
  />
);

export default SecretDetails;
