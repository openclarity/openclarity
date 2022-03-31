import React from 'react';
import classnames from 'classnames';

import './button.scss';

const Button = ({children, secondary, tertiary, disabled, onClick, className}) => (
    <button
        className={classnames(
            "ag-button",
            className,
            {"ag-button--primary": !secondary && !tertiary},
            {"ag-button--secondary": secondary},
            {"ag-button--tertiary": tertiary},
            {clickable: !!onClick}
        )}
        disabled={disabled}
        onClick={event => !disabled && onClick ? onClick(event) : null}
    >
        {children}
    </button>
)

export default Button;