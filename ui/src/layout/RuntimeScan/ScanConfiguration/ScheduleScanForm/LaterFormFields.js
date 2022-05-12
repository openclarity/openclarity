import React from 'react';
import moment from 'moment';
import { DateField, TimeField, validators } from 'components/Form';
import { formatFormFields, GENERAL_FOMR_FIELDS } from './utils';

const CLEAN_FORM_FIELDS = {
    LATER_DATE: "laterDate",
    LATER_TIME: "laterTime"
}

const FORM_FIELDS = formatFormFields(CLEAN_FORM_FIELDS);

export const convertServerToFormData = ({ScheduleScanConfigType, operationTime}) => {
    const selectedDate = moment(operationTime);

    return {
        [GENERAL_FOMR_FIELDS.SCHEDULE_TYPE]: ScheduleScanConfigType,
        [CLEAN_FORM_FIELDS.LATER_DATE]: !!operationTime ? selectedDate.format("YYYY-MM-DD") : "",
        [CLEAN_FORM_FIELDS.LATER_TIME]: !!operationTime ? selectedDate.format("HH:mm") : ""
    }
}

export const convertFormDataToServer = formValues => ({
    ScheduleScanConfigType: formValues[GENERAL_FOMR_FIELDS.SCHEDULE_TYPE],
    operationTime: moment(`${formValues[CLEAN_FORM_FIELDS.LATER_DATE]} ${formValues[CLEAN_FORM_FIELDS.LATER_TIME]}`).toISOString()
})

const LaterFormFields = () => (
    <div className="schedule-later-form-fields">
        <DateField name={FORM_FIELDS.LATER_DATE} label="Date" validate={validators.validateRequired} />
        <TimeField name={FORM_FIELDS.LATER_TIME} label="Time" validate={validators.validateRequired} />
    </div>
)

export default LaterFormFields;