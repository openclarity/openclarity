import React from 'react';
import classnames from 'classnames';

import './step-display.scss';

const StepDisplay = ({step, title, children, className, customContent: CustomContent}) => (
    <div className={classnames("step-display-wrapper", className)}>
        <div className="step-display">
            <div className="step-number">{step}</div>
            <div className="step-title">{title}</div>
            <div className="step-content">{children}</div>
        </div>
        {!!CustomContent && <CustomContent />}
    </div>
)

export default StepDisplay;