import React from 'react';
import { useLocation, useParams } from 'react-router-dom';
import BackRouteButton from 'components/BackRouteButton';
import Loader from 'components/Loader';
import TopBarTitle from 'components/TopBarTitle';
import { useFetch } from 'hooks';

import './details-page-wrapper.scss';

const DetailsPageWrapper = ({title, backTitle, url, getUrl, getReplace, detailsContent: DetailsContent}) => {
    const {pathname} = useLocation();
    const params = useParams();
    const {id} = params;
    
    const [{loading, data, error}, fetchData] = useFetch(!!url ? `${url}/${id}` : getUrl(params));

    return (
        <div className="details-page-wrapper">
            <TopBarTitle title={title} onRefresh={fetchData} loading={loading} />
            <div className="details-page-content-wrapper">
                <BackRouteButton title={backTitle} pathname={pathname.replace(!!getReplace ? getReplace(params) : `/${id}`, "")} />
                {loading ? <Loader /> : (!!error ? null : <DetailsContent data={data} />)}
            </div>
        </div>
    )
}

export default DetailsPageWrapper;