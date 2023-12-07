import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';

import './text-field.scss';

const TextField = ({className, label, disabled, tooltipText, type="text", placeholder, ...props}) => {
    const [field, meta] = useField(props);
    
    return (
        <div className={classnames("form-field-wrapper", "text-field", className)}>
            {!!label && <FieldLabel tooltipId={`form-tooltip-${field.name}`} tooltipText={tooltipText}>{label}</FieldLabel>}
            <input type={type} {...field} className="form-field" disabled={disabled} placeholder={placeholder} />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default TextField;