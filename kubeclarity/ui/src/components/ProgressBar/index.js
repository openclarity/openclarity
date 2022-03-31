import React from 'react';
import classnames from 'classnames';

import './progress-bar.scss';

const ProgressBar = ({percent=0, hidePercent=false, isSmall=false}) => (
    <div className="progress-bar-wrapper">
        <div className={classnames("progress-bar-container", {"is-small": isSmall})}>
            <div className={classnames("progress-bar-filler", {done: percent === 100})} style={{width: `${percent}%`}}></div>  
        </div>
        {!hidePercent && <div className="progress-bar-title">{`${percent}%`}</div>}
    </div>
)

export default ProgressBar;