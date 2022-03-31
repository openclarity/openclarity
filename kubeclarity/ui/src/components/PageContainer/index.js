import React from 'react';
import classnames from 'classnames';

import './page-container.scss';

const PageContainer = ({children, className, withPadding=false}) => (
    <div className={classnames("page-container", className, {"with-padding": withPadding})}>
        {children}
    </div>
);

export default PageContainer;