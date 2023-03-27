import React from 'react';
import classnames from 'classnames';
import Title from 'components/Title';

import './widget-wrapper.scss';

const WidgetWrapper = ({className, title, children}) => (
    <div className={classnames("dashboard-widget-wrapper", className)}>
        <Title removeMargin medium>{title}</Title>
        {children}
    </div>
)

export default WidgetWrapper;