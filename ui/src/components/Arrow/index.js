import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './arrow.scss';

export const ARROW_NAMES = {
    TOP: "top",
    BOTTOM: "bottom",
    RIGHT: "right",
    LEFT: "left"
}

const Arrow = ({name=ARROW_NAMES.TOP, onClick, disabled, small=false}) => {
    if (!Object.values(ARROW_NAMES).includes(name)) {
        console.error(`Arrow name '${name}' does not exist`);
    }

    return (
        <Icon
            name={ICON_NAMES.ARROW_HEAD_LEFT}
            className={classnames("arrow-icon", `${name}-arrow`, {small}, {clickable: !!onClick})}
            onClick={onClick}
            disabled={disabled}
            size={small ? 10 : 16}
        />
    );
}

export default Arrow;