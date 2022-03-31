import React from 'react';
import classnames from 'classnames';

import './title.scss';

const Title = ({children, small}) => (
    <div className={classnames("main-title", {small})}>{children}</div>
);

export default Title;