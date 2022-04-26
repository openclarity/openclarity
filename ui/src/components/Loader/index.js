import React from 'react';
import { SpinnerCircularFixed } from 'spinners-react';

import COLORS from 'utils/scss_variables.module.scss';

import './loader.scss';

const Loader = ({large=false, small=false}) => (
    <SpinnerCircularFixed
        className="ag-loader"
        size={large ? 78 : 64}
        thickness={large ? 100 : 50}
        speed={120}
        color={COLORS[large ? "color-main" : "color-main-light"]}
        secondaryColor={large ? "transparent" : COLORS["color-grey-light"]}
    />
);

export default Loader;