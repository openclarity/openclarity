import {mapValues, keyBy} from 'lodash';

export const OPERATORS = {
    is: {value: "is", label: "is"},
    isNot: {value: "isNot", label: "is not"},
    start: {value: "start", label: "starts with"},
    end: {value: "end", label: "ends with"},
    contains: {value: "contains", label: "contains"},
    gte: {value: "gte", label: "greater than or equal to"},
    lte: {value: "lte", label: "less than or equal to"},
    containElements: {value: "containElements", label: "contain elements"},
    doesntContainElements: {value: "doesntContainElements", label: "doesn't contain elements"},
}

export const formatFiltersToQueryParams = (filters) => {
    const filtersList = filters.map(({scope, operator, value} )=> ({key: `${scope}[${operator}]`, value}));

    return mapValues(keyBy(filtersList, "key"), "value");
};

export const getValueLabel = (valueItems=[], value) => {
    const valueItem = valueItems.find(valueItem => valueItem.value === value);
    
    return !!valueItem ? valueItem.label : value;
};