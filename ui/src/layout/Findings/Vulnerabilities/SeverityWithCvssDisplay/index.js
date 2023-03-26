import React from 'react';
import SeverityDisplay, { SEVERITY_ITEMS } from 'components/SeverityDisplay';
import { TooltipWrapper } from 'components/Tooltip';

import './severity-with-cvss-display.scss';

const SEVERITY_NOT_ALIGNED_MESSAGE = (
    <div style={{width: "330px"}}>
        While the scores may differ, the VMClarity (linux distribution severity) is more significant in your context than the CVSS base impact score.
    </div>
);

const SeverityWithCvssDisplay = ({severity, cvssScore, cvssSeverity, compareTooltipId}) => {
    if (!severity) {
        return null;
    }
    
    const showSeverityTooltip = !!cvssSeverity && cvssSeverity !== severity &&
        !(cvssSeverity === SEVERITY_ITEMS.NONE.value && severity === SEVERITY_ITEMS.NEGLIGIBLE.value);
    
    return (
        <div className="severity-with-cvss-display">
            <SeverityDisplay severity={severity} />
            {!!cvssScore && <div className="cvss-score-display">{`(cvss ${cvssScore})`}</div>}
            {!!showSeverityTooltip && <TooltipWrapper tooltipId={compareTooltipId} tooltipText={SEVERITY_NOT_ALIGNED_MESSAGE}>*</TooltipWrapper>}
        </div>
    );
}

export default SeverityWithCvssDisplay;