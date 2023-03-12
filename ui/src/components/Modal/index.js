import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import CloseButton from 'components/CloseButton';
import Button from 'components/Button';
import Title from 'components/Title';

import './modal.scss';

const Modal = (props) => {
    const {title, isMediumTitle=false, children, onClose, className, height=380, width=720, stickLeft=false, onDone, doneTitle="Done", disableDone=false, hideCancel=false,
        hideSubmit=false} = props;

    const [portalContainer, setPortalContainer] = useState(null);

    useEffect(() => {
        const container = document.querySelector("#main-wrapper");

        if (!container) {
            return;
        }
        
        setPortalContainer(container);
    }, []);

    if (!portalContainer) {
        return null;
    }

    return ReactDOM.createPortal(
        <div className="modal-outer-wrapper" onClick={event => {
            event.stopPropagation();
            event.preventDefault();

            onClose();
        }}>
            <div
                className={classnames("modal-inner-wrapper", {"stick-left": stickLeft}, className)}
                style={stickLeft ? {width: `${width}px`} : {height: `${height}px`, width: `${width}px`}}
                onClick={(event) => event.stopPropagation()}
            >
                <Title className="modal-title" medium={isMediumTitle}>{title}</Title>
                <div className="modal-content">{children}</div>
                <CloseButton onClose={onClose} />
                <div className="modal-actions">
                    {!hideCancel && <Button tertiary onClick={onClose}>Cancel</Button>}
                    {!hideSubmit && <Button className="modal-submit-button" onClick={onDone} disabled={disableDone}>{doneTitle}</Button>}
            </div>
            </div>
        </div>,
        portalContainer
    );
}

export default Modal;