import React from 'react';
import classnames from 'classnames';
import { useField } from 'formik';
import InfoIcon from 'components/InfoIcon';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';

import './checkbox-field.scss';

const CheckboxField = ({title, className, label, disabled, tooltipText, ...props}) => {
    const [field, meta, helpers] = useField(props);
    const {name} = field; 
    const {value} = meta;
    const {setValue} = helpers;

    const tooltipId = `form-tooltip-${name}`;
    
    return (
        <div className={classnames("form-field-wrapper", "checkbox-field-wrapper", className)}>
            {!!label && <FieldLabel tooltipId={tooltipId} tooltipText={tooltipText}>{label}</FieldLabel>}
            <label className={classnames("checkbox-wrapper", {disabled})}>
                <div className="inner-checkbox-wrapper">
                    <input type="checkbox" checked={value} value={value} name={name} onChange={event => disabled ? null : setValue(event.target.checked)} />
                    <span className="checkmark"></span>
                </div>
                <span className="checkbox-title">{title}</span>
                {!label && !!tooltipText && <div style={{marginLeft: "5px"}}><InfoIcon tooltipId={tooltipId} tooltipText={tooltipText} /></div>}
            </label>
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default CheckboxField;