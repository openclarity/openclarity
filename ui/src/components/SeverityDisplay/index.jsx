import React from 'react';
import Icon from 'components/Icon';
import { toCapitalized } from 'utils/utils';
import { VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';

import './severity-display.scss';

import COLORS from 'utils/scss_variables.module.scss';

export const SEVERITY_ITEMS = {
    CRITICAL: {valueKey: "CRITICAL", color: COLORS["color-error-dark"]},
    HIGH: {valueKey: "HIGH", color: COLORS["color-error"]},
    MEDIUM: {valueKey: "MEDIUM", color: COLORS["color-warning"]},
    LOW: {valueKey: "LOW", color: COLORS["color-warning-low"]},
    NEGLIGIBLE: {valueKey: "NEGLIGIBLE", color: COLORS["color-status-blue"]},
    NONE: {valueKey: "NONE", color: COLORS["color-status-blue"]}
}

const SeverityDisplay = ({severity, score}) => {
    const {color} = SEVERITY_ITEMS[severity];

    return (
        <div className="severity-display">
            {!!score ? <div style={{color}}>{score}</div> : <Icon name={VULNERABIITY_FINDINGS_ITEM.icon} size={25} style={{color}} />}
            <div className="severity-title" style={!!score ? {color} : undefined}>{toCapitalized(severity)}</div>
        </div>
    )
}

export default SeverityDisplay;