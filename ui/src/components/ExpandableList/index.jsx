import React, { useState } from 'react';
import classnames from 'classnames';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import Tag from 'components/Tag';

import './expandable-list.scss';
import { ValueWithFallback } from 'components/ValueWithFallback';

const MINIMIZED_LEN = 1;

const ExpandableList = ({items, withTagWrap=false}) => {
    const allItems = items || [];
    const minmizedItems = allItems.slice(0, MINIMIZED_LEN);

    const [itemsToDisplay, setItemsToDisplay] = useState(allItems.length > MINIMIZED_LEN ? minmizedItems : allItems);
    const isOpen = itemsToDisplay.length === allItems.length;

    return (
        <div>
            <div className="expandable-list-display-wrapper">
                <div className="expandable-list-items">
                    {
                        <ValueWithFallback>
                            {itemsToDisplay.map((item, index) => (
                                <div key={index} className="expandable-list-item-wrapper">
                                    <div className={classnames("expandable-list-item")}>
                                        {withTagWrap ? <Tag>{item}</Tag> : item}
                                    </div>
                                </div>
                            ))}
                        </ValueWithFallback>
                    }
                </div>
                {minmizedItems.length !== allItems.length &&
                    <Arrow
                        name={isOpen ? ARROW_NAMES.TOP : ARROW_NAMES.BOTTOM}
                        onClick={event => {
                            event.stopPropagation();
                            event.preventDefault();
                            
                            setItemsToDisplay(isOpen ? minmizedItems : allItems);
                        }}
                        small
                    />
                }
            </div>
        </div>
    )
}

export default ExpandableList;
