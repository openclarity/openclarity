import React, { useEffect } from 'react';
import { useField } from 'formik';

import { RadioButtonGroup, TextField, useFormikContext, validators } from 'components/Form';
import { CustomQueryBuilder } from './CustomQueryBuilder';

import { SCOPE_CONFIG_INITIAL_VALUE } from "./ScanConfigWizardModal.constants";

const RADIO_BUTTONS = [
    { label: "All", value: true },
    { label: "Define scope", value: false }
];

const StepGeneralProperties = () => {
    const { values: formValues } = useFormikContext();
    const { scanTemplate } = formValues;

    const [scopeConfigField, , scopeConfigHelpers] = useField("scanTemplate.scopeConfig");
    const { setValue: setScopeConfigValue } = scopeConfigHelpers;

    const [, , scopeHelpers] = useField("scanTemplate.scope");
    const { setValue: setScopeValue } = scopeHelpers;

    const [, , fullScopeHelpers] = useField("scanTemplate.fullScope");
    const { setValue: setFullScopeValue } = fullScopeHelpers;

    useEffect(() => {
        if (scanTemplate.fullScope) {
            setScopeValue("");
            setScopeConfigValue(SCOPE_CONFIG_INITIAL_VALUE);
        }
        // eslint-disable-next-line
    }, [scanTemplate.fullScope])

    useEffect(() => {
        if (scopeConfigField?.value && Object.keys(scopeConfigField.value).length > 2) {
            setFullScopeValue(false);
        }
        // eslint-disable-next-line
    }, [])

    return (
        <div className="scan-config-general-step">
            <TextField
                name="name"
                label="Scan config name*"
                placeholder="Type a scan config name..."
                validate={validators.validateRequired}
            />
            <RadioButtonGroup items={RADIO_BUTTONS} label="Scope*" name="scanTemplate.fullScope" initialValue={RADIO_BUTTONS[0].value} />
            {
                scanTemplate.fullScope === RADIO_BUTTONS[1].value &&
                <CustomQueryBuilder />
            }
        </div>
    )
}

export default StepGeneralProperties;
