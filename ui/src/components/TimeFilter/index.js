import React from 'react';
import moment from 'moment';
import DateTimeSelect from 'components/DateTimeSelect';
import DropdownSelect from 'components/DropdownSelect';

import './time-filter.scss';

export const TIME_SELECT_ITEMS = {
    MIN: {
        value: "MIN",
        label: "Last 5 minutes",
        calc: () => ({endTime: moment().toISOString(), startTime: moment().subtract(5, 'minutes').toISOString()})
    },
    HOUR: {
        value: "HOUR",
        label: "Last hour",
        calc: () => ({endTime: moment().toISOString(), startTime: moment().subtract(1, 'hours').toISOString()})
    },
    DAY: {
        value: "DAY",
        label: "Last day",
        calc: () => ({endTime: moment().toISOString(), startTime: moment().subtract(1, 'days').toISOString()})
    },
    WEEK: {
        value: "WEEK",
        label: "Last week",
        calc: () => ({endTime: moment().toISOString(), startTime: moment().subtract(7, 'days').toISOString()})
    },
    MONTH: {
        value: "MONTH",
        label: "Last month",
        calc: () => ({endTime: moment().toISOString(), startTime: moment().subtract(1, 'months').toISOString()})
    },
    CUSTOM: {
        value: "CUSTOM",
        label: "Custom"
    }
}

export const getNow = () => moment().toISOString();

export const getTimeFormat = (startTime, endTime) => {
    const daysDiff = moment(endTime).diff(moment(startTime), 'days', true);
    const yearsDiff = moment(endTime).diff(moment(startTime), 'years', true);

    if (daysDiff < 1) {
        return "HH:mm:ss";
    }
    if (yearsDiff < 1) {
        return "MMM Do, HH:mm"
    }

    return "MMM Do, YYYY HH:mm";
};

const CustomDateRange = ({startTime, endTime, onChange}) => (
    <React.Fragment>
        <DateTimeSelect
            value={startTime}
            onChange={startTime => onChange({startTime, endTime})}
            maxDate={endTime}
        />
        <DateTimeSelect
            value={endTime}
            onChange={endTime => onChange({startTime, endTime})}
            minDate={startTime}
            maxDate={moment().toDate()}
            isEnd={true}
        />
    </React.Fragment>
);

const TimeFilter = ({selectedRange=TIME_SELECT_ITEMS.HOUR.value, startTime, endTime, onChange}) => {
    const isCustom = selectedRange === TIME_SELECT_ITEMS.CUSTOM.value;
    const selectItems = Object.values(TIME_SELECT_ITEMS);
    const selectedValue = selectItems.find(item => item.value === selectedRange);

    return (
        <div className="time-filter-wrapper">
            {isCustom &&
                <CustomDateRange
                    startTime={moment(startTime).toDate()}
                    endTime={moment(endTime).toDate()}
                    onChange={({endTime, startTime}) =>
                        onChange({selectedRange, startTime: moment(startTime).toISOString(), endTime: moment(endTime).toISOString()})}
                />}
            <DropdownSelect
                items={selectItems}
                value={selectedValue}
                onChange={selectedRangeItem => {
                    const {value} = selectedRangeItem;
                    const {calc} = TIME_SELECT_ITEMS[value];

                    if (!!calc) {
                        onChange({selectedRange: value, ...calc()})
                    } else {
                        onChange({selectedRange: value, startTime: getNow(), endTime: getNow()})
                    }
                }}
                small
            />
        </div>
    );
}

export default TimeFilter;