import React, { useState } from 'react';
import classnames from 'classnames';
import moment from 'moment';
import { isNull } from 'lodash';
import { SingleDatePicker } from 'react-dates';
import 'react-dates/initialize';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import Arrow from 'components/Arrow';
import { FieldLabel, FieldError } from 'components/Form/utils';

import 'react-dates/lib/css/_datepicker.css';
import './date-field.scss';

const DATE_FORMAT = 'YYYY-MM-DD';

const DateField = (props) => {
    const {label, className, tooltipText, displayFormat="MMM Do"} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    const [focused, setFocused] = useState(false);

    const formattedValue = !!value ? moment(value) : null;

    return (
        <div className={classnames("form-field-wrapper", "date-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <SingleDatePicker
                date={formattedValue}
                onDateChange={date => setValue(isNull(date) ? "" : moment(date).format(DATE_FORMAT))}
                focused={focused}
                onFocusChange={({focused}) => setFocused(focused)}
                id={name}
                daySize={30}
                numberOfMonths={1}
                hideKeyboardShortcutsPanel={true}
                small={true}
                displayFormat={displayFormat}
                navPrev={<Arrow name="left" />}
                navNext={<Arrow name="right" />}
                placeholder="Date"
                readOnly={true}
            />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default DateField;