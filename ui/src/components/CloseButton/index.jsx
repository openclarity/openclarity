import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';

import './close-button.scss';

const CloseButton = ({onClose, small=false}) => (
    <Icon name={ICON_NAMES.X_MARK} onClick={onClose} className="close-button" size={small ? 14 : 22} />
)

export default CloseButton;