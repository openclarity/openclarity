import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './close-button.scss';

const CloseButton = ({onClose, small=false}) => (
    <Icon name={ICON_NAMES.X_MARK} onClick={onClose} className={classnames("close-button", {small})} />
)

export default CloseButton;