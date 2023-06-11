import { isEmpty, isNumber } from 'lodash';

export const validateRequired = value => (
    isEmpty(value) && !isNumber(value) && value !== 0 ? "This field is required" : undefined
);

export const validateMaxScanField = value => (
    value < 0 ? "Should be greater than or equal to 1" : undefined   
);