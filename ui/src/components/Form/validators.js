import { isEmpty, isNumber } from 'lodash';

export const validateRequired = value => (
    isEmpty(value) && !isNumber(value) && value !== 0 ? "This field is required" : undefined
);