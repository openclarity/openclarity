import React from 'react';
import { isEmpty } from 'lodash';
import classnames from 'classnames';
import TimePicker from 'react-time-picker';
import { useField } from 'formik';
import { FieldLabel, FieldError } from 'components/Form/utils';

import './time-field.scss';

const TimeField = (props) => {
    const {label, className, tooltipText} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    return (
        <div className={classnames("form-field-wrapper", "time-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <TimePicker onChange={setValue} value={value} format="HH:mm" clearIcon={null} clockIcon={null} disableClock />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default TimeField;