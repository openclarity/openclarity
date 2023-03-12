
import React from 'react';
import Button from 'components/Button';
import Icon from 'components/Icon';

import './button-with-icon.scss';

const ButtonWithIcon = ({onClick, iconName, disabled, children}) => (
    <Button onClick={onClick} className="button-with-icon" disabled={disabled}>
        <Icon name={iconName} size={20} />
        <span>{children}</span>
    </Button>
);

export default ButtonWithIcon;