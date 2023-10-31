
import React from 'react';
import { isEmpty } from 'lodash';

import './value-with-fallback.scss';

export const ValueWithFallback = ({
    children,
    fallback = <span className="value-with-fallback-empty">-</span>
}) => {
    return isEmpty(children) ? fallback : <>{children}</>;
}
