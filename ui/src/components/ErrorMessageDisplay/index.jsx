import React from 'react';

import './error-message-display.scss';

const ErrorMessageDisplay = ({children, title}) => (
    <div className="error-message-display">
        {!!title && <div className="error-message-display-title">{title}</div>}
        <div className="error-message-display-message">{children}</div>
    </div>
)

export default ErrorMessageDisplay;