import React, { useEffect } from 'react';
import classnames from 'classnames';
import { useFetch } from 'hooks';
import PageContainer from 'components/PageContainer';
import Loader from 'components/Loader';
import WidgetTitle from '../WidgetTitle';

import './widget-wrapper.scss';

const WidgetWrapper = ({url, title, widget: Widget, className, refreshTimestamp}) => {
    const [{loading, data, error}, fetchData] = useFetch(url);

    useEffect(() => {
        fetchData();
    }, [fetchData, refreshTimestamp]);

    return (
        <PageContainer className={classnames("dashboard-widget-wrapper", className)} withPadding>
            <WidgetTitle>{title}</WidgetTitle>
            {loading ? <Loader /> : (error ? <div>Error loading data</div> : <Widget data={data} />)}
        </PageContainer>
    )
}

export default WidgetWrapper;