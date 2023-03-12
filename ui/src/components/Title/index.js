import React from 'react';
import classnames from 'classnames';

import './title.scss';

const Title = ({children, className, medium, onClick, removeMargin=false}) => (
    <div className={classnames("clarity-title", className, {clickable: !!onClick}, {medium}, {"no-margin": removeMargin})} onClick={onClick}>
        {children}
    </div>
)

export default Title;