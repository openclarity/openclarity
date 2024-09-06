import React from "react";
import classnames from "classnames";
import { useLocation, useParams } from "react-router-dom";
import BackRouteButton from "components/BackRouteButton";
import ContentContainer from "components/ContentContainer";
import Loader from "components/Loader";
import Title from "components/Title";

import "./details-page-wrapper.scss";

const DetailsContentWrapper = ({
  data,
  getTitleData,
  detailsContent: DetailsContent,
}) => {
  const { title, subTitle } = getTitleData(data);

  return (
    <div className="details-page-content-wrapper">
      <div className="details-page-title">
        <Title removeMargin>{title}</Title>
        {!!subTitle && <div className="details-page-title-sub">{subTitle}</div>}
      </div>
      <ContentContainer>
        <DetailsContent data={data} />
      </ContentContainer>
    </div>
  );
};

const DetailsPageWrapper = ({
  className,
  backTitle,
  getTitleData,
  detailsContent,
  data,
  isError,
  isLoading,
  withPadding = false,
}) => {
  const { pathname } = useLocation();
  const params = useParams();
  const { id } = params;
  const innerPath = params["*"];

  return (
    <div
      className={classnames("details-page-wrapper", className, {
        "with-padding": withPadding,
      })}
    >
      <BackRouteButton
        title={backTitle}
        pathname={pathname.replace(
          !!innerPath ? `/${id}/${innerPath}` : `/${id}`,
          "",
        )}
      />
      {isLoading && <Loader />}
      {!isLoading && !isError && (
        <DetailsContentWrapper
          detailsContent={detailsContent}
          getTitleData={getTitleData}
          data={data}
        />
      )}
    </div>
  );
};

export default DetailsPageWrapper;
