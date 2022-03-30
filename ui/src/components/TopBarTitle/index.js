import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';

import './top-bar-title.scss';

const TopBarTitle = ({title, loading, onRefresh}) => (
    <div className="top-bar-title-wrapper">
        <div className="top-bar-title">{`KUBECLARITY ${!!title ? `| ${title}` : ""}`}</div>
        {!!onRefresh && !loading && <Icon name={ICON_NAMES.REFRESH} onClick={onRefresh} />}
    </div>
)

export default TopBarTitle;