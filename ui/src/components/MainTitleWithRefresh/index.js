import React from 'react';
import Title from 'components/Title';
import Icon, { ICON_NAMES } from 'components/Icon';

import './main-title-with-refresh.scss';

const MainTitleWithRefresh = ({title, onRefreshClick}) => (
    <div className="main-title-with-refresh">
        <Title>{title}</Title>
        <Icon name={ICON_NAMES.REFRESH} onClick={onRefreshClick} />
    </div>
)

export default MainTitleWithRefresh;