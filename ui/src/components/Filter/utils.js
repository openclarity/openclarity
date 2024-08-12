export const OPERATORS = {
    eq: {value: "eq", label: "is"},
    ne: {value: "ne", label: "is not"},
    startswith: {value: "startswith", label: "starts with"},
    endswith: {value: "endswith", label: "ends with"},
    contains: {value: "contains", label: "contains"},
    notcontains: {value: "notcontains", label: "don't contain"},
    gt: {value: "gt", label: "greater than"},
    ge: {value: "ge", label: "greater than or equal to"},
    lt: {value: "lt", label: "less than"},
    le: {value: "le", label: "less than or equal to"}
}

const SPECIAL_CASE_OPERATORS = [
    OPERATORS.contains.value,
    OPERATORS.startswith.value,
    OPERATORS.endswith.value
];

export const formatFiltersToOdataItems = (filters) => {
    return filters.reduce((acc, curr) => {
        const {scope, operator, value, isNumber, isDate, customOdataFormat} = curr;
        const valuesList = Array.isArray(value) ? value : [value];
        
        const formatValueItem = valueItem => isNumber ? valueItem : `'${isDate ? (new Date(valueItem)).toISOString() : valueItem}'`;
        const formatItem = valueItem => {
            if (SPECIAL_CASE_OPERATORS.includes(operator)) {
                return `${operator}(${scope},${formatValueItem(valueItem)})`;
            }
            
            return `${scope} ${operator} ${formatValueItem(valueItem)}`;
        }

        return [
            ...acc,
            !!customOdataFormat ? customOdataFormat(valuesList, operator, scope) :`(${valuesList.map(valueItem => formatItem(valueItem)).join(` or `)})`
        ];
    }, []);
};

export const getValueLabel = (valueItems=[], value) => {
    const valueItem = valueItems.find(valueItem => valueItem.value === value);
    
    return !!valueItem ? valueItem.label : value;
};