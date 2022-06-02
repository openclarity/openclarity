import React, { useEffect } from 'react';
import classnames from 'classnames';
import { pickBy, isInteger } from 'lodash';
import { usePrevious } from 'hooks';
import { SelectField, TimeField, TextField, useFormikContext, validators } from 'components/Form';
import { formatFormFields, SCHEDULE_TYPE_DATA_WRAPPER, SCHEDULE_TYPES, GENERAL_FOMR_FIELDS } from './utils';

const CLEAN_FORM_FIELDS = {
    REPETITIVE_TYPE: "repetitiveType",
    TIME_OF_DAY: "timeOfDay",
    WEEK_DAY: "dayInWeek",
    HOURS_INTERVAL: "hoursInterval",
    DAYS_INTERVAL: "daysInterval",
    WEEKS_INTERVAL: "weeksInterval" //display only
}

const FORM_FIELDS = formatFormFields(CLEAN_FORM_FIELDS);

const REPETITIVE_INTERVAL_TYPES = {
    ByHoursScheduleScanConfig: {value: "ByHoursScheduleScanConfig", label: "Hours", intervalKey: CLEAN_FORM_FIELDS.HOURS_INTERVAL},
    ByDaysScheduleScanConfig: {value: "ByDaysScheduleScanConfig", label: "Days", intervalKey: CLEAN_FORM_FIELDS.DAYS_INTERVAL},
    WeeklyScheduleScanConfig: {value: "WeeklyScheduleScanConfig", label: "Weeks", intervalKey: CLEAN_FORM_FIELDS.WEEKS_INTERVAL}
}

const REPETITIVE_WEEK_DAYS = {
    SUNDAY: {value: 1, label: "Sunday"},
    MONDAY: {value: 2, label: "Monday"},
    TUESDAY: {value: 3, label: "Tuesday"},
    WEDNSDAY: {value: 4, label: "Wednsday"},
    THURSDAY: {value: 5, label: "Thursday"},
    FRIDAY: {value: 6, label: "Friday"},
    SATURDAY: {value: 7, label: "Saturday"}
}

const formatTimeDigit = time => ("0" + time).slice(-2);

export const convertServerToFormData = ({ScheduleScanConfigType, timeOfDay, ...data}) => {
    const repetitiveTypeIsSet = Object.keys(REPETITIVE_INTERVAL_TYPES).includes(ScheduleScanConfigType);
    const {hour, minute} = timeOfDay || {};

    return {
        [GENERAL_FOMR_FIELDS.SCHEDULE_TYPE]: SCHEDULE_TYPES.REPETITIVE.value,
        [CLEAN_FORM_FIELDS.REPETITIVE_TYPE]: repetitiveTypeIsSet ? ScheduleScanConfigType : REPETITIVE_INTERVAL_TYPES.ByHoursScheduleScanConfig.value,
        [CLEAN_FORM_FIELDS.TIME_OF_DAY]: !!timeOfDay ? `${formatTimeDigit(hour)}:${formatTimeDigit(minute)}` : "",
        [CLEAN_FORM_FIELDS.WEEK_DAY]: data[CLEAN_FORM_FIELDS.WEEK_DAY] || "",
        [CLEAN_FORM_FIELDS.HOURS_INTERVAL]: data[CLEAN_FORM_FIELDS.HOURS_INTERVAL] || "",
        [CLEAN_FORM_FIELDS.DAYS_INTERVAL]: data[CLEAN_FORM_FIELDS.DAYS_INTERVAL] || "",
        [CLEAN_FORM_FIELDS.WEEKS_INTERVAL]: 1
    }
}

export const convertFormDataToServer = formValues => {
    const [hour, minute] = formValues[CLEAN_FORM_FIELDS.TIME_OF_DAY].split(":");

    const ScheduleScanConfigType = formValues[CLEAN_FORM_FIELDS.REPETITIVE_TYPE];
    const {intervalKey} = REPETITIVE_INTERVAL_TYPES[ScheduleScanConfigType];

    const submitData = {
        ScheduleScanConfigType,
        dayInWeek: formValues[CLEAN_FORM_FIELDS.WEEK_DAY],
        ...(!!hour ? {timeOfDay: {hour: parseInt(hour), minute: parseInt(minute)}} : {}),
        ...(ScheduleScanConfigType === REPETITIVE_INTERVAL_TYPES.WeeklyScheduleScanConfig.value ? {} : {[intervalKey]: formValues[intervalKey]})
    };
    
    return pickBy(submitData, (value) => value !== "");
}

const RepeatTimeField = ({label, className}) => (
    <TimeField className={classnames("repetitive-time-field", className)} name={FORM_FIELDS.TIME_OF_DAY} label={label} validate={validators.validateRequired} />
)

const RepetitiveFormFields = () => {
    const {values, setFieldValue} = useFormikContext();
    const repetitiveType = values[SCHEDULE_TYPE_DATA_WRAPPER][CLEAN_FORM_FIELDS.REPETITIVE_TYPE];
    const prevRepetitiveType = usePrevious(repetitiveType)
    const {intervalKey} = REPETITIVE_INTERVAL_TYPES[repetitiveType] || {};

    const isHoursInerval = repetitiveType === REPETITIVE_INTERVAL_TYPES.ByHoursScheduleScanConfig.value;
    const isDaysInerval = repetitiveType === REPETITIVE_INTERVAL_TYPES.ByDaysScheduleScanConfig.value;
    const isWeeklyInerval = repetitiveType === REPETITIVE_INTERVAL_TYPES.WeeklyScheduleScanConfig.value;

    useEffect(() => {
        if (!prevRepetitiveType || prevRepetitiveType === repetitiveType) {
            return;
        }

        if (isHoursInerval) {
            setFieldValue(FORM_FIELDS.TIME_OF_DAY, "");
        } else {
            setFieldValue(FORM_FIELDS.WEEK_DAY, "");
        }

        if (!isWeeklyInerval) {
            setFieldValue(intervalKey, "");
        }
    }, [prevRepetitiveType, repetitiveType, isHoursInerval, intervalKey, isWeeklyInerval, setFieldValue]);

    return (
        <div className="repetitive-later-form-fields">
            <TextField
                className="repetitive-interval-field"
                type="number"
                name={`${SCHEDULE_TYPE_DATA_WRAPPER}.${intervalKey}`}
                label="Repeat every"
                validate={value => {
                    const requiredError = validators.validateRequired(value);
                    
                    return !!requiredError ? requiredError : ((isInteger(value) && value > 0) ? null : "Invalid value");
                }}
                disabled={isWeeklyInerval}
            />
            <SelectField
                className="repetitive-type-field"
                name={FORM_FIELDS.REPETITIVE_TYPE}
                items={Object.values(REPETITIVE_INTERVAL_TYPES)}
                validate={validators.validateRequired}
            />
            {isDaysInerval && <RepeatTimeField label="On" />}
            {isWeeklyInerval &&
                <React.Fragment>
                    <SelectField
                        className="repetitive-weekday-field"
                        name={FORM_FIELDS.WEEK_DAY}
                        label="On"
                        items={Object.values(REPETITIVE_WEEK_DAYS)}
                        validate={validators.validateRequired}
                    />
                    <RepeatTimeField className="with-top-margin" />
                </React.Fragment>
            }
        </div>
    )
}

export default RepetitiveFormFields;