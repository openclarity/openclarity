import { isEmpty, isNumber } from 'lodash';

export const validateRequired = value => (
    isEmpty(value) && !isNumber(value) && value !== 0 ? "This field is required" : undefined
);

export const keyValueListValidator = (valuesList) => {
    const errorItem = valuesList.find(item => {
        const valueList = item.split("=");

        return valueList.length !== 2 || valueList.find(item => item === "");
    })

    return !!errorItem ? "Values must be in the key=value format" : null;
}