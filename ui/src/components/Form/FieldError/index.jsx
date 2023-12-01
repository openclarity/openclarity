import React from 'react';

import './field-error.scss';

const FieldError = ({children}) => (
    <div className="form-field-error">{children}</div>
);

export default FieldError;
