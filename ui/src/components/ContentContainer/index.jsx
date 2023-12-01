import React from 'react';
import classnames from 'classnames';

import './content-container.scss';

const ContentContainer = ({children, withMargin=false}) => (
    <div className={classnames("content-container-wrapper", {"with-margin": withMargin})}>
        {children}
    </div>
);

export default ContentContainer;