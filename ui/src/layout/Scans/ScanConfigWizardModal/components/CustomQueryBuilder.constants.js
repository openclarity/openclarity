import { BasicConfig } from "@react-awesome-query-builder/ui";
import { AntdWidgets } from '@react-awesome-query-builder/antd';

const { FieldCascader } = AntdWidgets;

const BASIC_OPERATORS = {
    equal: {
        //...BasicConfig.operators.equal,
        label: 'equals',
        labelForFormat: 'eq',
        // formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
        //     const opStr = isForDisplay ? "=" : opDef.label;
        //     if (valueTypes == "boolean" && isForDisplay)
        //         return value == "No" ? `NOT ${field}` : `${field}`;
        //     else
        //         return `${field} ${opStr} ${value}`;
        // },
    },
    not_equal: {
        // ...BasicConfig.operators.not_equal,
        label: 'does not equal',
        labelForFormat: 'ne',
    },
    greater: {
        //...BasicConfig.operators.greater,
        label: 'greater than',
        labelForFormat: 'gt',
    },
    greater_or_equal: {
        //...BasicConfig.operators.greater_or_equal,
        label: 'greater than or equal',
        labelForFormat: 'ge',
    },
    less: {
        //...BasicConfig.operators.less,
        label: 'less than',
        labelForFormat: 'lt',
    },
    less_or_equal: {
        //...BasicConfig.operators.less_or_equal,
        label: 'less than or equal',
        labelForFormat: 'le',
    },
}

const BETWEEN_OPERATORS = {
    between: {
        ...BasicConfig.operators.between,
        label: "in between",
        labelForFormat: 'in betweeen',
        formatOp: (field, op, values, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay) => {
            let valFrom = values.first();
            let valTo = values.get(1);
            if (isForDisplay)
                return `${field} BETWEEN ${valFrom} AND ${valTo}`;
            else
                return `${field} ge ${valFrom} and ${field} le ${valTo}`;
        },

    },
    not_between: {
        ...BasicConfig.operators.not_between,
        label: "not in between",
        labelForFormat: 'not in between',
        formatOp: (field, op, values, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay) => {
            let valFrom = values.first();
            let valTo = values.get(1);
            if (isForDisplay)
                return `${field} NOT BETWEEN ${valFrom} AND ${valTo}`;
            else
                return `(${field} lt ${valFrom} or ${field} gt ${valTo})`;
        },
    }
}

const FUNCTION_OPERATORS = {
    ends_with: {
        ...BasicConfig.operators.ends_with,

    },
    starts_with: {
        ...BasicConfig.operators.starts_with,
    },
    multiselect_contains: {
        ...BasicConfig.operators.multiselect_contains,
    },
    //    ...BasicConfig.operators
}

const IS_NULL_NOT_NULL = {
    is_null: {
        ...BasicConfig.operators.is_null,
        label: "is null",
        formatOp: (field, op, value, valueSrc, valueType, opDef, operatorOptions, isForDisplay) => {
            return isForDisplay ? `${field} IS NULL` : `${field} is null`;
        },
    },
    is_not_null: {
        ...BasicConfig.operators.is_not_null,
        label: "is not null",
        formatOp: (field, op, value, valueSrc, valueType, opDef, operatorOptions, isForDisplay) => {
            return isForDisplay ? `${field} IS NOT NULL` : `${field} is not null`;
        },
    }
}

const OPERATORS = {
    //...BasicConfig.operators,
    ...BASIC_OPERATORS,
    ...BETWEEN_OPERATORS,
    ...IS_NULL_NOT_NULL
}

const BASIC_OPERATORS_LIST = Object.keys(BASIC_OPERATORS);
const FUNCTION_OPERATORS_LIST = Object.keys(FUNCTION_OPERATORS);
const BETWEEN_OPERATORS_LIST = Object.keys(BETWEEN_OPERATORS);

const BASIC_CONFIG = {
    ...BasicConfig,
    settings: {
        ...BasicConfig.settings,
        renderField: (props) => <FieldCascader {...props} />,
        fieldSeparator: "/"
    },
    conjunctions: {
        ...BasicConfig.conjunctions,
        AND: {
            ...BasicConfig.conjunctions.AND,
            label: 'and',
            formatConj: (children, conj, not, isForDisplay) => {
                return children.size > 1
                    ? (not ? "not " : "") + "(" + children.join(" " + (isForDisplay ? "AND" : "and") + " ") + ")"
                    : (not ? "not (" : "") + children.first() + (not ? ")" : "");
            },
        },
        OR: {
            ...BasicConfig.conjunctions.OR,
            label: 'or',
            formatConj: (children, conj, not, isForDisplay) => {
                return children.size > 1
                    ? (not ? "not " : "") + "(" + children.join(" " + (isForDisplay ? "OR" : "or") + " ") + ")"
                    : (not ? "not (" : "") + children.first() + (not ? ")" : "");
            },
        },
    },
    operators: {
        ...OPERATORS
    },
    types: {
        ...BasicConfig.types,
        'group-select': {
            ...BasicConfig.types.select,
            valueSources: [], //['value', 'field', 'func'],
            defaultOperator: 'equal',
            widgets: {
                'group-select': {
                    operators: ['is_null', 'is_not_null'],
                },
            },
        },
        number: {
            ...BasicConfig.types.number,
            valueSources: [], //['value', 'field', 'func'],
            widgets: {
                number: {
                    operators: BASIC_OPERATORS_LIST
                },
            },
        },
        datetime: {
            ...BasicConfig.types.datetime,
            widgets: {
                datetime: {
                    operators: [...BASIC_OPERATORS_LIST, ...BETWEEN_OPERATORS_LIST],
                },
            },
        }
        // array: {
        //     ...BasicConfig.types['!struct'],
        //valueSources: [], //['value', 'field', 'func'],
        //defaultOperator: 'equal',
        // widgets: {
        //     'group-select': {
        //         operators: ['equal', 'not_equal'],
        //     },
        // },
        //},
    },
    widgets: {
        ...BasicConfig.widgets,
        'group-select': {
            ...BasicConfig.widgets.select,
        },
    },
    fieldSources: ["field", "func"],
    fields: {}
};

export { BASIC_CONFIG }; 