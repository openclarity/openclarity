import React from 'react';
import classnames from 'classnames';
import { useLocation, useParams } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import ContentContainer from 'components/ContentContainer';
import Loader from 'components/Loader';
import Title from 'components/Title';
import { useFetch } from 'hooks';

import './details-page-wrapper.scss';

const DetailsContentWrapper = ({data, getTitleData, detailsContent: DetailsContent, fetchData}) => {
    const {title, subTitle} = getTitleData(data);

    return (
        <div className="details-page-content-wrapper">
            <div className="details-page-title">
                <Title removeMargin>{title}</Title>
                {!!subTitle && <div className="details-page-title-sub">{subTitle}</div>}
            </div>
            <ContentContainer><DetailsContent data={data} fetchData={fetchData} /></ContentContainer>
        </div>
    )
}

const DetailsPageWrapper = ({className, backTitle, url, expand, select, getTitleData, detailsContent, withPadding=false}) => {
    const {pathname} = useLocation();
    const params = useParams();
    const {id} = params;
    const innerPath = params["*"];
    
    const expandParams = !!expand ? `$expand=${expand}` : "";
    const selectParams = !!select ? `$select=${select}` : "";
    const [{loading, data, error}, fetchData] = useFetch(
        `${url}/${id}${!!expandParams || !!selectParams ? "?" : ""}${selectParams}${!!selectParams ? "&" : ""}${expandParams}`
    );

    return (
        <div className={classnames("details-page-wrapper", className, {"with-padding": withPadding})}>
            <BackRouteButton title={backTitle} pathname={pathname.replace(!!innerPath ? `/${id}/${innerPath}` : `/${id}`, "")} />
            {loading ? <Loader /> : (!!error ? null :
                <DetailsContentWrapper detailsContent={detailsContent} getTitleData={getTitleData} data={data} fetchData={fetchData} />)
            }
        </div>
    )
}

export default DetailsPageWrapper;