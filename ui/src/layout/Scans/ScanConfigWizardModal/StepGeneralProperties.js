import React from 'react';
import { TextField, validators } from 'components/Form';

const StepGeneralProperties = () => {
    return (
        <div className="scan-config-general-step">
            <TextField
                name="name"
                label="Scan config name*"
                placeholder="Type a name..."
                validate={validators.validateRequired}
            />
            <TextField
                name="scanTemplate.scope"
                label="Scope"
                placeholder="Type an ODATA $filter to reduce assets to scan..."
            />
        </div>
    )
}

export default StepGeneralProperties;
