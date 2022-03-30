import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import CloseButton from 'components/CloseButton';
import Button from 'components/Button';

import './modal.scss';

const Modal = (props) => {
    const {title, children, onClose, className, height=380, width=720, stickLeft=false, onDone, doneTitle="Done", disableDone=false, hideCancel=false,
        hideSubmit=false} = props;

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

    return ReactDOM.createPortal(
        <div className="modal-outer-wrapper">
            <div className={classnames("modal-inner-wrapper", {"stick-left": stickLeft}, className)} style={stickLeft ? {} : {height: `${height}px`, width: `${width}px`}}>
                <div className="modal-title">{title}</div>
                <div className="modal-content">{children}</div>
                <CloseButton onClose={onClose} />
                <div className="modal-actions">
                    {!hideCancel && <Button tertiary onClick={onClose}>Cancel</Button>}
                    {!hideSubmit && <Button onClick={onDone} disabled={disableDone}>{doneTitle}</Button>}
            </div>
            </div>
        </div>,
        portalContainer
    );
}

export default Modal;