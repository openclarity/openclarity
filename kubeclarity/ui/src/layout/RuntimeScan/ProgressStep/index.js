import React, { useState } from 'react';
import { isEmpty } from 'lodash';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import ProgressBar from 'components/ProgressBar';
import Button from 'components/Button';
import CloseButton from 'components/CloseButton';
import StepDisplay from '../StepDisplay';

import './progress-step.scss';

const ProgressErrorDisplay = ({erorrs}) => {
    const [showDetails, setShowDetails] = useState(false);

    return (
        <div className="progress-error-display-wrapper">
            <div className="progress-error-inticator">
                <Icon name={ICON_NAMES.ALERT} />
                <div>Some of the elements were failed to be scanned.</div>
                <Button tertiary className="progress-error-details-link" onClick={() => setShowDetails(true)}>Click here for more details</Button>
            </div>
            {showDetails &&
                <div className="progress-error-details">
                    <CloseButton onClose={() => setShowDetails(false)} small />
                    <div className="progress-error-message">
                        {erorrs.map((scanError, index) => <div key={index}>{scanError}</div>)}
                    </div>
                </div>
            }
        </div>
    )
}

const ProgressStep = ({title, isDone, percent, scanErrors}) => {
    const hasErrors = !isEmpty(scanErrors);
    
    return (
        <StepDisplay step="1"  title="Progress:" className="progress-step-display" customContent={hasErrors && (() => <ProgressErrorDisplay erorrs={scanErrors} />)}>
            <div className="progress-wrapper">
                <ProgressBar percent={percent} />
                <div className="progress-status-wrapper">
                    <div className="progress-status">{title}</div>
                    {isDone &&
                        <Icon name={hasErrors ? ICON_NAMES.ALERT : ICON_NAMES.CHECK_MARK} className={classnames("status-icon", {"is-error": hasErrors})} />
                    }
                </div>
            </div>
        </StepDisplay>
    )
}

export default ProgressStep;