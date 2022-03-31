import React from 'react';
import { useNavigate } from 'react-router-dom';
import classnames from 'classnames';

import './inner-app-link.scss';

const InnerAppLink = ({pathname, children, onClick, className, withUnderline=true}) => {
    const navigate = useNavigate();

    const handleOnClick = event => {
        event.preventDefault();
        event.stopPropagation();

        if (!!onClick) {
            onClick();
        }
        
        navigate(pathname);
    }

    return (
        <div className="inner-app-link-wrapper">
            <div className={classnames("inner-app-link", {underline: withUnderline}, className)} onClick={handleOnClick}>{children}</div>
        </div>
    )
}

export default InnerAppLink;