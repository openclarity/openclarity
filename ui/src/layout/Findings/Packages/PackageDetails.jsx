import React from "react";
import { useLocation } from "react-router-dom";
import TabbedPage from "components/TabbedPage";
import FindingsDetailsPage from "../FindingsDetailsPage";
import TabPackageDetails from "./TabPackageDetails";
import AssetsForFindingTable from "layout/Assets/AssetsForFindingTable";

const PACKAGE_DETAILS_PATHS = {
  PACKAGE_DETAILS: "",
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
          title: "Package details",
          isIndex: true,
          path: PACKAGE_DETAILS_PATHS.PACKAGE_DETAILS,
          component: () => <TabPackageDetails data={data} />,
        },
        {
          id: "assets",
          title: "Assets",
          path: PACKAGE_DETAILS_PATHS.ASSET_LIST,
          component: () => <AssetsForFindingTable findingId={id} />,
        },
      ]}
      withInnerPadding={false}
    />
  );
};

const PackageDetails = () => (
  <FindingsDetailsPage
    backTitle="Packages"
    getTitleData={({ findingInfo }) => ({
      title: findingInfo.name,
      subTitle: findingInfo.version,
    })}
    detailsContent={DetailsContent}
  />
);

export default PackageDetails;
