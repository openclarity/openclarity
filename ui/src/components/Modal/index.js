import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import classnames from 'classnames';
import CloseButton from 'components/CloseButton';
import Button from 'components/Button';
import Title from 'components/Title';

import './modal.scss';

const WIDE_MODAL_STYLE = {
    minWidth: "720px",
    width: "75%",
    maxWidth: "1200px"
}

const Modal = (props) => {
    const {
        children,
        className,
        disableDone = false,
        doneTitle = "Done",
        height = 380,
        hideCancel = false,
        hideSubmit = false,
        isMediumTitle = false,
        onClose,
        onDone,
        removeTitleMargin = false,
        stickLeft = false,
        title,
        wideModal = false,
        width = 720,
    } = props;

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
                className={classnames("modal-inner-wrapper", { "stick-left": stickLeft }, className)}
                style={{
                    ...stickLeft ? { width: `${width}px` } : { height: `${height}px`, width: `${width}px` },
                    ...wideModal ? WIDE_MODAL_STYLE : null
                }}
                onClick={(event) => event.stopPropagation()}
            >
                <Title className="modal-title" medium={isMediumTitle} removeMargin={removeTitleMargin}>{title}</Title>
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