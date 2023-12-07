import React from 'react';
import classnames from 'classnames';

import './tag.scss';
import { ValueWithFallback } from 'components/ValueWithFallback';

const Tag = ({children, onClick}) => (
    <div className={classnames("clarity-tag", {clickable: !!onClick})} onClick={onClick}>{children}</div>
)

export const TagsList = ({items}) => (
    <div className="clarity-tags-list">
        <ValueWithFallback>
            {items?.map((tag, index) => <Tag key={index}>{tag}</Tag>)}
        </ValueWithFallback>
    </div>
)

export default Tag;
