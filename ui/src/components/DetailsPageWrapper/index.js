import React from 'react';
import classnames from 'classnames';
import { useLocation, useParams } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import ContentContainer from 'components/ContentContainer';
import Loader from 'components/Loader';
import Title from 'components/Title';
import { useFetch } from 'hooks';

import './details-page-wrapper.scss';

const DetailsContentWrapper = ({data, getTitleData, detailsContent: DetailsContent}) => {
    const {title, subTitle} = getTitleData(data);

    return (
        <div className="details-page-content-wrapper">
            <div className="details-page-title">
                <Title removeMargin>{title}</Title>
                {!!subTitle && <div className="details-page-title-sub">{subTitle}</div>}
            </div>
            <ContentContainer><DetailsContent data={data} /></ContentContainer>
        </div>
    )
}

const DetailsPageWrapper = ({className, backTitle, url, getUrl, getReplace, getTitleData, detailsContent}) => {
    const {pathname} = useLocation();
    const params = useParams();
    const {id} = params;
    
    const [{loading, data, error}] = useFetch(!!url ? `${url}/${id}` : getUrl(params));

    return (
        <div className={classnames("details-page-wrapper", className)}>
            <BackRouteButton title={backTitle} pathname={pathname.replace(!!getReplace ? getReplace(params) : `/${id}`, "")} />
            {loading ? <Loader /> : (!!error ? null : <DetailsContentWrapper detailsContent={detailsContent} getTitleData={getTitleData} data={data} />)}
        </div>
    )
}

export default DetailsPageWrapper;