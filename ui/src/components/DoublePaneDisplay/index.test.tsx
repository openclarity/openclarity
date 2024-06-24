import React from 'react';
import { describe, test, expect } from 'vitest'
import { render } from '@testing-library/react';
import DoublePaneDisplay from './index';

describe('DoublePaneDisplay', () => {
    test('renders without error', () => {
        const wrapper = render(<DoublePaneDisplay className="" rightPlaneDisplay="" leftPaneDisplay="" />);
        expect(wrapper).toBeTruthy();

        const left = wrapper.container.querySelector('.left-pane-display');
        expect(left).toBeTruthy();

        const right = wrapper.container.querySelector('.right-pane-display');
        expect(right).toBeTruthy();
    });
});
