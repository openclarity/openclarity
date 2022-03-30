import React from 'react';
import { SEVERITY_ITEMS } from 'utils/systemConsts';

import './severity-tag.scss';

const SeverityTag = ({severity}) => {
    const {label, color} = SEVERITY_ITEMS[severity] || {};

    return (
        <div className="severity-tag" style={{backgroundColor: color}}>{label || severity}</div>
    )
}

export default SeverityTag;