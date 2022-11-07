import React from 'react';
import classnames from 'classnames';
import Icon from 'components/Icon';

export const NAMESPACES_SELECT_TITLE = "Select the target namespaces to scan";
export const NAMESPACES_SELECT_INFO_MESSAGE = "Select specific namespaces to scan or leave empty to scan the whole K8s cluster. After removing namespace(s), click the 'Start scan' button ";

export const ConfigButton = ({children, icon, disabled, onClick}) => (
    <div className={classnames("runtime-scan-config-button", {disabled})} onClick={onClick}>
        <Icon name={icon} />
        <span>{children}</span>
    </div>
)
