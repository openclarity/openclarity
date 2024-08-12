import React from 'react';

import './field-set.scss';

export const FieldSet = ({ legend, children}) =>
    <fieldset>
        <legend>{legend}</legend>
        {children}
    </fieldset>
