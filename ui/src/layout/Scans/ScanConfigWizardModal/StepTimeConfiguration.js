import React, { useEffect } from 'react';
import { SelectField, DateField, TimeField, useFormikContext, validators } from 'components/Form';
import { usePrevious } from 'hooks';

export const SCHEDULE_TYPES_ITEMS = {
    NOW: {value: "NOW", label: "Now"},
    LATER: {value: "LATER", label: "Specified time"}
}

const LaterFormFields = () => {
    return (
        <div className="schedule-later-form-fields">
            <DateField
                name="scheduled.laterDate"
                label="Date*"
                validate={validators.validateRequired}
            />
            <TimeField
                name="scheduled.laterTime"
                label="Time*"
                validate={validators.validateRequired}
            />
        </div>
    )
}

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
        </div>
    )
}

export default StepTimeConfiguration;