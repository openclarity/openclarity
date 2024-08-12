import React from 'react';
import { useNavigate } from 'react-router-dom';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './back-route-button.scss';

const BackRouteButton = ({title, pathname}) => {
    const navigate = useNavigate();

    return (
        <div className="back-route-button" onClick={() => navigate(pathname)}>
            <Arrow name={ARROW_NAMES.LEFT} />
            <div className="back-title">{title}</div>
        </div>
    )
}

export default BackRouteButton;