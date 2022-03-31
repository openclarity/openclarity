import React from 'react';
import classnames from 'classnames';

import noResultsImage from 'utils/images/empty.svg';

import './no-results-display.scss';

const NoResultsDisplay = ({title, isSmall=false}) => {
    return (
        <div className="empty-results-display-wrapper">
            <div className="no-results-title">{title}</div>
            <img src={noResultsImage} alt="no results" className={classnames("no-results-image", {"is-small": isSmall})} />
        </div>
    )
}

export default NoResultsDisplay;