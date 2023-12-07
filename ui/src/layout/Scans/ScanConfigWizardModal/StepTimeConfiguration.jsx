import React, { useEffect } from 'react';
import { SelectField, DateField, TimeField, CronField, useFormikContext, validators } from 'components/Form';
import { usePrevious } from 'hooks';

import 'react-js-cron/dist/styles.css';

export const SCHEDULE_TYPES_ITEMS = {
    NOW: {value: "NOW", label: "Now"},
    LATER: {value: "LATER", label: "Specified time"},
    REPETITIVE: {value: "REPETITIVE", label: "Repetitive"}
}

export const CRON_QUICK_OPTIONS = [
    {value: "0 */12 * * *", label: "Every 12 hours"},
    {value: "0 0 * * *", label: "Once a day"},
    {value: "0 0 * * 5", label: "Once a week"}
];

const LaterFormFields = () => (
    <div className="schedule-later-form-fields">
        <DateField
            name="scheduled.laterDate"
            label="Date*"
            validate={validators.validateRequired}
            minDate={new Date()}
        />
        <TimeField
            name="scheduled.laterTime"
            label="Time*"
            validate={validators.validateRequired}
        />
    </div>
)

const RepetitiveFormFields = () => (
    <div className="repetitive-later-form-fields">
        <CronField
            name="scheduled.cronLine"
            quickOptions={CRON_QUICK_OPTIONS}
        />
    </div>
)

const StepTimeConfiguration = () => {
    const {values, validateForm} = useFormikContext();
    
    const {scheduledSelect} = values?.scheduled;
    const prevScheduledSelect = usePrevious(scheduledSelect);

    useEffect(() => {
        if (prevScheduledSelect !== scheduledSelect) {
            validateForm();
        }
    }, [prevScheduledSelect, scheduledSelect, validateForm]);

    return (
        <div className="scan-config-time-config-step">
            <SelectField
                name="scheduled.scheduledSelect"
                items={Object.values(SCHEDULE_TYPES_ITEMS)}
            />
            {scheduledSelect === SCHEDULE_TYPES_ITEMS.LATER.value &&
                <LaterFormFields />
            }
            {scheduledSelect === SCHEDULE_TYPES_ITEMS.REPETITIVE.value &&
                <RepetitiveFormFields />
            }
        </div>
    )
}

export default StepTimeConfiguration;
