import React from 'react';
import classnames from 'classnames';

import './double-pane-display.scss';

const DoublePaneDisplay = ({className, rightPlaneDisplay: RightPlaneDisplay, leftPaneDisplay: LeftPaneDisplay}) => (
    <div className={classnames("double-pane-display-wrapper", className)}>
        <div className="left-pane-display">{!!LeftPaneDisplay && <LeftPaneDisplay />}</div>
        <div className="right-pane-display">{!!RightPlaneDisplay && <RightPlaneDisplay />}</div>
    </div>
)

export default DoublePaneDisplay;