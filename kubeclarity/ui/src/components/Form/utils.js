import React from 'react';
import InfoIcon from 'components/InfoIcon';

export const FieldLabel = ({children, tooltipText, tooltipId}) => (
    <div className="field-label-wrapper">
        <label className="form-field-label">{children}</label>
        {!!tooltipText && <InfoIcon tooltipId={tooltipId} tooltipText={tooltipText} />}
    </div>
);

export const FieldError = ({children}) => (
    <div className="form-field-error">{children}</div>
)