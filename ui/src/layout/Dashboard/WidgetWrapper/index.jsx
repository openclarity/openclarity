import React from 'react';
import classnames from 'classnames';
import Title from 'components/Title';

import './widget-wrapper.scss';

const WidgetWrapper = ({className, title, children, titleMargin=20}) => (
    <div className={classnames("dashboard-widget-wrapper", className)}>
        <div style={{marginBottom: `${titleMargin}px`}}><Title removeMargin medium>{title}</Title></div>
        {children}
    </div>
)

export default WidgetWrapper;