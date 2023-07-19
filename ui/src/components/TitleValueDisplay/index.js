import React, { useState } from 'react';
import classnames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';

import './title-value-display.scss';

export const TitleValueDisplayRow = ({children}) => (
    <div className="title-value-display-row">{children}</div>
);

export const TitleValueDisplayColumn = ({children}) => (
    <div className="title-value-display-column">{children}</div>
);

export const ValuesListDisplay = ({values}) => (
    <div className="title-value-values-list">
        {values?.map((value, index) => <div key={index} className="title-value-values-list-item">{value}</div>)}
    </div>
)

const TitleValueDisplay = ({title, children, className, withOpen=false, defaultOpen=false, isSubItem=false}) => {
    const [isOpen, setIsOpen] = useState(defaultOpen);

    return (
        <div className={classnames("title-value-display-wrapper", className, {"sub-item": isSubItem})}>
            <div className={classnames("title-value-display-title-wrapper", {"with-open": withOpen})} onClick={() => setIsOpen(!isOpen)}>
                <div className="title-value-display-title">{title}</div>
                {withOpen && <Arrow name={isOpen ? ARROW_NAMES.TOP : ARROW_NAMES.BOTTOM} small />}
            </div>
            {(!withOpen || isOpen) && <div className="title-value-display-content">{children}</div>}
        </div>
    );
}

export default TitleValueDisplay
