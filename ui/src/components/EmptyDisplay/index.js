import React from 'react';
import { ICON_NAMES } from 'components/Icon';
import ButtonWithIcon from 'components/ButtonWithIcon';
import Button from 'components/Button';

import emptyImage from 'utils/images/empty.svg';

import './empty-display.scss';

const EmptyDisplay = ({message, title, onClick, subTitle, onSubClick}) => (
    <div className="empty-display-wrapper">
        <div className="empty-display">
            <div className="empty-dispaly-title">{message}</div>
            <img src={emptyImage} alt="no-data" />
            <div className="empty-display-buttons">
                {subTitle && <Button tertiary onClick={onSubClick}>{subTitle}</Button>}
                <ButtonWithIcon onClick={onClick} iconName={ICON_NAMES.PLUS}>{title}</ButtonWithIcon>
            </div>
        </div>
    </div>
)

export default EmptyDisplay;