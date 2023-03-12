import React from 'react';

import './tag.scss';

const Tag = ({children}) => (
    <div className="clarity-tag">{children}</div>
)

export const TagsList = ({items}) => (
    <div className="clarity-tags-list">{items.map((tag, index) => <Tag key={index}>{tag}</Tag>)}</div>
)

export default Tag;