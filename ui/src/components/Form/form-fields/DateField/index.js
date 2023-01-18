import React from 'react';
import classnames from 'classnames';
import { isNull } from 'lodash';
import DateTimePicker from 'react-datetime-picker';
import { isEmpty } from 'lodash';
import { useField } from 'formik';
import Arrow from 'components/Arrow';
import { FieldLabel, FieldError } from 'components/Form/utils';
import { formatDateBy } from 'utils/utils';

import './date-field.scss';

const DATE_FORMAT = 'YYYY-MM-DD';

const DateField = (props) => {
    const {label, className, tooltipText, displayFormat="MMM dd"} = props;
    const [field, meta, helpers] = useField(props);
    const {name, value} = field; 
    const {setValue} = helpers;

    const formattedValue = !!value ? new Date(value) : null;

    return (
        <div className={classnames("form-field-wrapper", "date-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <DateTimePicker
                onChange={date => setValue(isNull(date) ? "" : formatDateBy(date, DATE_FORMAT))}
                value={formattedValue}
                className="date-field-select"
                calendarClassName="date-field-select-calendar"
                calendarIcon={null}
                clearIcon={null}
                disableClock={true}
                format={displayFormat}
                name={name}
                openWidgetsOnFocus={true}
                minDate={new Date()}
                prevLabel={<Arrow name="left" />}
                nextLabel={<Arrow name="right" />}
                prev2Label={null}
                next2Label={null}
                minDetail="month"
                calendarType="US"
            />
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default DateField;