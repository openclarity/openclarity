import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import CloseButton from 'components/CloseButton';

import './notification.scss';

export const NOTIFICATION_TYPES = {
    INFO: "INFO",
    ERROR: "ERROR"
}

const Notification = ({message, type, onClose}) => {
    const [portalContainer, setPortalContainer] = useState(null);

    useEffect(() => {
        const container = document.querySelector("main");

        if (!container) {
            return;
        }
        
        setPortalContainer(container);
    }, []);

    if (!portalContainer) {
        return null;
    }

    const notificationType = type || NOTIFICATION_TYPES.INFO; 

    return ReactDOM.createPortal(
        <div className={classnames("notification-wrapper", notificationType.toLowerCase())}>
            <div className="notification-content">{message}</div>
            <CloseButton onClose={onClose} small />
        </div>,
        portalContainer
    );
}

export default Notification;