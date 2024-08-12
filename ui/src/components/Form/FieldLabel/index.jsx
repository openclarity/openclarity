import React from 'react';
import InfoIcon from 'components/InfoIcon';

import './field-label.scss';

const FieldLabel = ({children, tooltipText, tooltipId}) => (
    <div className="form-field-label-wrapper">
        <label className="form-field-label">{children}</label>
        {!!tooltipText && <InfoIcon tooltipId={tooltipId} tooltipText={tooltipText} />}
    </div>
);

export default FieldLabel;