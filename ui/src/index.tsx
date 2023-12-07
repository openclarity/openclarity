import React from 'react';
import { createRoot } from 'react-dom/client';
import App from 'layout/App';

import 'utils/fonts/fonts.scss';

const container = document.getElementById('root');
if (container === null) {
    throw new Error('Root element not found');
}
const root = createRoot(container);

root.render(<App />);
