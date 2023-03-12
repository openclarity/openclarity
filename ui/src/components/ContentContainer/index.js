import React from 'react';

import './content-container.scss';

const ContentContainer = ({children}) => (
    <div className="content-container-wrapper">
        {children}
    </div>
);

export default ContentContainer;