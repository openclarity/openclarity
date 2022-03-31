import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFetch } from 'hooks';
import Icon, { ICON_NAMES } from 'components/Icon';
import { ROUTES } from 'utils/systemConsts';

import COLORS from 'utils/scss_variables.module.scss';

import './counters-display.scss';

const CounterBox = ({count, title, icon, color, loading, onClick}) => (
    <div className="dashboard-counter-box" style={{backgroundColor: color}} onClick={onClick}>
        <Icon name={icon} />
        <div className="counter-box-data">
            <div className="counter-box-count">{loading ? "-" : (count || 0)}</div>
            <div className="counter-box-title">{title}</div>
        </div>
    </div>
)

const CountersDisplay = ({refreshTimestamp}) => {
    const navigate = useNavigate();

    const [{loading, data}, fetchData] = useFetch("dashboard/counters");

    useEffect(() => {
        fetchData();
    }, [fetchData, refreshTimestamp]);

    const {applications, resources, packages} = data || {};

    return (
        <div className="dashboard-counters-wrapper">
            <CounterBox
                count={applications}
                title="Applications"
                icon={ICON_NAMES.APPLICATION}
                color={COLORS["color-main-dark"]}
                loading={loading}
                onClick={() => navigate(ROUTES.APPLICATIONS)}
            />
            <CounterBox
                count={resources}
                title="Application resources"
                icon={ICON_NAMES.RESOURCE}
                color={COLORS["color-main"]}
                loading={loading}
                onClick={() => navigate(ROUTES.APPLICATION_RESOURCES)}
            />
            <CounterBox
                count={packages}
                title="Packages"
                icon={ICON_NAMES.PACKAGE}
                color={COLORS["color-dash-blue-light"]}
                loading={loading}
                onClick={() => navigate(ROUTES.PACKAGES)}
            />
        </div>
    )
}

export default CountersDisplay;