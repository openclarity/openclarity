import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';
import { toCapitalized } from 'utils/utils';

import './score-tag.scss';

const SCORE_ITEMS = {
    NONE: {value: "NONE"},
    HIGH: {value: "HIGH", icon: ICON_NAMES.ARROW_UP},
    LOW: {value: "LOW", icon: ICON_NAMES.ARROW_UP, style: {transform: "rotate(180deg)"}}
}

const ScoreTag = ({score}) => {
    const {icon, style} = SCORE_ITEMS[score] || {};

    return (
        <div className="score-tag">
            <div>{toCapitalized(score)}</div>
            {!!icon && <Icon name={icon} size={14} style={style} />}
        </div>
    )
}

export default ScoreTag;