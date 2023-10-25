import React from 'react';
import { RadioButtonGroup, TextField, useFormikContext, validators } from 'components/Form';
import { CustomQueryBuilder } from './components';

const RADIO_BUTTONS = [
    { label: "All", value: "all" },
    { label: "Define scope", value: "define-scope" }
];

const INITIAL_QUERY = {
    combinator: 'and',
    rules: [],
};

const StepGeneralProperties = () => {
    const { values: formValues } = useFormikContext();
    const { scanTemplate } = formValues;

    return (
        <div className="scan-config-general-step">
            <TextField
                name="name"
                label="Scan config name*"
                placeholder="Type a scan config name..."
                validate={validators.validateRequired}
            />
            <RadioButtonGroup items={RADIO_BUTTONS} label="Scope*" name="scanTemplate.scopeSelector" initialValue={RADIO_BUTTONS[0].value} />
            {
                scanTemplate.scopeSelector === RADIO_BUTTONS[1].value &&
                <CustomQueryBuilder initialQuery={INITIAL_QUERY} name="scanTemplate.scope" />
            }
        </div>
    )
}

export default StepGeneralProperties;
