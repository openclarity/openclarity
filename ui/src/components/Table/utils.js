import React from 'react';
import Icon, { ICON_NAMES } from 'components/Icon';

export const EmptyValue = () => (
    <Icon name={ICON_NAMES.STROKE} className="table-empty-value" />
);

export const StatusIndicatorIcon = ({isSuccess}) => (
    isSuccess ? <Icon name={ICON_NAMES.CHECK_MARK} className="status-indicator-icon" /> : <EmptyValue />
);