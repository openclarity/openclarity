import React from 'react';
import classnames from 'classnames';

import './button.scss';

const Button = ({children, secondary, tertiary, disabled, onClick, className}) => (
    <button
        type="button"
        className={classnames(
            "clarity-button",
            className,
            {"clarity-button--primary": !secondary && !tertiary},
            {"clarity-button--secondary": secondary},
            {"clarity-button--tertiary": tertiary},
            {clickable: !!onClick}
        )}
        disabled={disabled}
        onClick={event => !disabled && onClick ? onClick(event) : null}
    >
        {children}
    </button>
)

export default Button;