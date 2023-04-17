
import React from 'react';
import classnames from 'classnames';
import Button from 'components/Button';
import Icon from 'components/Icon';

import './button-with-icon.scss';

const ButtonWithIcon = ({onClick, iconName, disabled, className, children}) => (
    <Button onClick={onClick} className={classnames("button-with-icon", className)} disabled={disabled}>
        <Icon name={iconName} size={20} />
        <span>{children}</span>
    </Button>
);

export default ButtonWithIcon;