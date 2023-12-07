import React from 'react';
import { useNavigate } from 'react-router-dom';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './links-list.scss';

const LinksList = ({items}) => {
    const navigate = useNavigate();

    const onItemClick = (path, callback) => {
        if (!!callback) {
            callback(path);
        }
        
        navigate(path);
    }

    return (
        <div className="links-list-wrapper">
            {
                items.map(({path, component: Component, callback}, index) => (
                    <div key={index} className="links-list-item" onClick={() => onItemClick(path, callback)}>
                        <div className="links-list-item-content"><Component /></div>
                        <Arrow name={ARROW_NAMES.RIGHT} />
                    </div>
                ))
            }
        </div>
    )
}

export default LinksList;