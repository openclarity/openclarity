import React from 'react';
import { useNavigate } from 'react-router-dom';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './links-list.scss';

const LinksList = ({items}) => {
    const navigate = useNavigate();

    return (
        <div className="links-list-wrapper">
            {
                items.map(({path, component: Component}, index) => (
                    <div key={index} className="links-list-item" onClick={() => navigate(path)}>
                        <div className="links-list-item-content"><Component /></div>
                        <Arrow name={ARROW_NAMES.RIGHT} />
                    </div>
                ))
            }
        </div>
    )
}

export default LinksList;