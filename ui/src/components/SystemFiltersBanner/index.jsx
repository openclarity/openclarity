import React from 'react';
import { useNavigate } from 'react-router-dom';
import CloseButton from 'components/CloseButton';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './system-filter-banner.scss';

const SystemFilterBanner = ({onClose, displayText, backPath, customDisplay: CustomDisplay}) => {
    const navigate = useNavigate();

    return (
        <div className="system-filter-banner">
            <div className="system-filter-content">
                <Arrow name={ARROW_NAMES.LEFT} small onClick={() => navigate(backPath)} />
                <div className="filter-content">{displayText}</div>
            </div>
            <div style={{display: "flex", alignItems: "center"}}>
                {!!CustomDisplay && <CustomDisplay />}
                <CloseButton small onClose={onClose} />
            </div>
        </div>
    )
}

export default SystemFilterBanner;