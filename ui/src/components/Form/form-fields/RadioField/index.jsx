import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';

import './radio-field.scss';

const RadioField = ({items, className, label, disabled, tooltipText, ...props}) => {
    const [field, meta, helpers] = useField(props);
    const {name} = field; 
    const {setValue} = helpers;
    
    return (
        <div className={classnames("form-field-wrapper", "radio-field-wrapper", className)}>
            {!!label && <FieldLabel tooltipId={`form-tooltip-${field.name}`} tooltipText={tooltipText}>{label}</FieldLabel>}
            {
                items.map(({value, label}) => (
                    <label key={value} className="radio-field-item">
                        <span className="radio-text">{label}</span>
                            <input
                                type="radio"
                                name={name}
                                checked={value === meta.value}
                                value={value}
                                onChange={() => setValue(value)}
                            />
                            <span className="checkmark"></span>
                    </label>
                ))
            }
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default RadioField;