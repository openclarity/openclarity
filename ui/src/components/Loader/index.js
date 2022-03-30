import React from 'react';
import { SpinnerCircularFixed } from 'spinners-react';

import COLORS from 'utils/scss_variables.module.scss';

import './loader.scss';

const Loader = () => (
    <SpinnerCircularFixed
        className="ag-loader"
        size={64}
        thickness={50}
        speed={120}
        color={COLORS["color-main-light"]}
        secondaryColor={COLORS["color-grey-light"]}
    />
);

export default Loader;