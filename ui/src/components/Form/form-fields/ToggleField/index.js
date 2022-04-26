import React from 'react';
import Toggle from 'react-toggle';
import classnames from 'classnames';
import { useField } from 'formik';

import 'react-toggle/style.css';
import './toggle-field.scss';

const ToggleField = (props) => {
    const {label, className, disabled} = props;
    const [field, meta, helpers] = useField(props);
    const {name} = field; 
    const {setValue} = helpers;
    
    return (
        <label className={classnames("form-field-wrapper", "toggle-field", {disabled}, {[className]: className})}>
            <div>{label}</div>
            <Toggle
                name={name}
                icons={false}
                checked={!meta.value}
                onChange={({target}) => setValue(!target.checked)} 
                disabled={disabled}
            />
        </label>
    )
}

export default ToggleField;