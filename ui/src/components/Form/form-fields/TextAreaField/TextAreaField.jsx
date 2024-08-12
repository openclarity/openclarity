import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';

import './TextAreaField.scss';

const TextAreaField = ({ className, label, disabled, tooltipText, placeholder, ...props }) => {
    const [field, meta] = useField(props);

    return (
        <div className={classnames("form-field-wrapper", "textarea-field", className)}>
            {!!label && <FieldLabel tooltipId={`form-tooltip-${field.name}`} tooltipText={tooltipText}>{label}</FieldLabel>}
            <textarea {...field} className="form-field" disabled={disabled} placeholder={placeholder} />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default TextAreaField;