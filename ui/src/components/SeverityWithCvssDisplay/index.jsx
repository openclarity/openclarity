import React from 'react';
import SeverityDisplay, { SEVERITY_ITEMS } from 'components/SeverityDisplay';
import { TooltipWrapper } from 'components/Tooltip';
import { toCapitalized } from 'utils/utils';

import './severity-with-cvss-display.scss';

export {
    SEVERITY_ITEMS
}

const SeverityWithCvssDisplay = ({severity, cvssScore, cvssSeverity, compareTooltipId}) => {
    if (!severity) {
        return null;
    }
    
    const showSeverityTooltip = !!cvssSeverity && cvssSeverity !== severity &&
        !(cvssSeverity === SEVERITY_ITEMS.NONE.value && severity === SEVERITY_ITEMS.NEGLIGIBLE.value);

    const severityNotAlignedMessage = (
        <div style={{width: "330px"}}>
            {`Although the CVSS base impact score is ${cvssScore} (${toCapitalized(cvssSeverity || "")}), the linux distribution severity reflects the risk more accurately.`}
        </div>
    );
    
    return (
        <div className="severity-with-cvss-display">
            <SeverityDisplay severity={severity} />
            {!!cvssScore && <div className="cvss-score-display">{`(cvss ${cvssScore})`}</div>}
            {!!showSeverityTooltip && <TooltipWrapper tooltipId={compareTooltipId} tooltipText={severityNotAlignedMessage}>*</TooltipWrapper>}
        </div>
    );
}

export default SeverityWithCvssDisplay;