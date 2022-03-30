import React from 'react';

const VerticalItemsList = ({items}) => {
    if (!items) {
        return null;
    }
    
    return (
        <div>{items.map((item, index) => <div key={index}>{item}</div>)}</div>
    )
}

export default VerticalItemsList;