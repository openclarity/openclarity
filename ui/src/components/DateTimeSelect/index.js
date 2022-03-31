import React from 'react';
import DateTimePicker from 'react-datetime-picker';
import { formatDateBy } from 'utils/utils';
import Icon, { ICON_NAMES } from 'components/Icon';

import './date-time-select.scss';

const DateTimeSelect = ({value, onChange, minDate, maxDate, isEnd}) => (
    <DateTimePicker
        onChange={onChange}
        value={value}
        className="date-time-select"
        calendarClassName="date-time-select-calendar"
        clearIcon={null}
        format="dd/MM/y HH:mm"
        disableClock={true}
        autoFocus={false}
        next2Label={null}
        maxDate={maxDate}
        minDate={minDate}
        formatShortWeekday={(locale, date) => formatDateBy(date, 'dd')}
        formatMonth={(locale, date) => formatDateBy(date, 'MMM')}
        hourPlaceholder="hh"
        minutePlaceholder="mm"
        dayPlaceholder="dd"
        monthPlaceholder="mm"
        yearPlaceholder="yyyy"
        calendarIcon={<Icon className="date-time-picket-icon" name={isEnd ? ICON_NAMES.END_TIME : ICON_NAMES.START_TIME} />}
    />
)

export default DateTimeSelect;