import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './filter-button.scss';

const FilterButton = ({onClick, pressed, children}) => (
    <button type="button" onClick={onClick} className={classnames("filter-button", {pressed})}>
        <Icon name={ICON_NAMES.FILTER} size={16} />
        <span>{children}</span>
    </button>
);

export default FilterButton;