import React from 'react';
import Toggle from 'react-toggle';

import 'react-toggle/style.css';
import './toggle-button.scss';

const ToggleButton = ({title, checked, onChange}) => (
    <div className="toggle-button">
        <Toggle
            icons={false}
            checked={checked}
            onChange={({target}) => onChange(target.checked)}
        />
        <div>{title}</div>
    </div>
)

export default ToggleButton;