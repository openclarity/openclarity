import { AntdConfig, AntdWidgets } from '@react-awesome-query-builder/antd';
const { FieldCascader } = AntdWidgets;

const CONJUNCTIONS = {
    ...AntdConfig.conjunctions,
    AND: {
        ...AntdConfig.conjunctions.AND,
        label: 'and',
        formatConj: (children, conj, not, isForDisplay) => {
            return children.size > 1
                ? (not ? "not " : "") + "(" + children.join(" " + (isForDisplay ? "AND" : "and") + " ") + ")"
                : (not ? "not (" : "") + children.first() + (not ? ")" : "");
        },
    },
    OR: {
        ...AntdConfig.conjunctions.OR,
        label: 'or',
        formatConj: (children, conj, not, isForDisplay) => {
            return children.size > 1
                ? (not ? "not " : "") + "(" + children.join(" " + (isForDisplay ? "OR" : "or") + " ") + ")"
                : (not ? "not (" : "") + children.first() + (not ? ")" : "");
        },
    },
}


const BASIC_OPERATORS = {
    equal: {
        ...AntdConfig.operators.equal,
        label: 'equal',
        labelForFormat: 'eq',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            const opStr = isForDisplay ? "=" : opDef.labelForFormat;
            if (typeof valueTypes === "boolean" && isForDisplay) {
                return value === "No" ? `NOT ${field}` : `${field}`;
            } else {
                return `${field} ${opStr} ${value}`;
            }
        },
    },
    not_equal: {
        ...AntdConfig.operators.not_equal,
        label: 'not equal to',
        labelForFormat: 'ne',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            if (typeof valueTypes === "boolean" && isForDisplay)
                return value === "No" ? `${field}` : `NOT ${field}`;
            else
                return `${field} ${isForDisplay ? "!=" : opDef.labelForFormat} ${value}`;
        },
    },
    greater: {
        ...AntdConfig.operators.greater,
        label: 'greater than',
        labelForFormat: 'gt',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            const opStr = isForDisplay ? ">" : opDef.labelForFormat;
            return `${field} ${opStr} ${value}`;
        },
    },
    greater_or_equal: {
        ...AntdConfig.operators.greater_or_equal,
        label: 'greater than or equal',
        labelForFormat: 'ge',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            const opStr = isForDisplay ? ">=" : opDef.labelForFormat;
            return `${field} ${opStr} ${value}`;
        }
    },
    less: {
        ...AntdConfig.operators.less,
        label: 'less than',
        labelForFormat: 'lt',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            const opStr = isForDisplay ? "<" : opDef.labelForFormat;
            return `${field} ${opStr} ${value}`;
        }
    },
    less_or_equal: {
        ...AntdConfig.operators.less_or_equal,
        label: 'less than or equal',
        labelForFormat: 'le',
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            const opStr = isForDisplay ? "<=" : opDef.labelForFormat;
            return `${field} ${opStr} ${value}`;
        }
    },
}

const BETWEEN_OPERATORS = {
    between: {
        ...AntdConfig.operators.between,
        label: "in between",
        labelForFormat: 'in betweeen',
        formatOp: (field, op, values, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay) => {
            let valFrom = values.first();
            let valTo = values.get(1);
            if (isForDisplay)
                return `(${field} BETWEEN ${valFrom} AND ${valTo})`;
            else
                return `${field} ge ${valFrom} and ${field} le ${valTo}`;
        },
    },
    not_between: {
        ...AntdConfig.operators.not_between,
        label: "not in between",
        labelForFormat: 'not in between',
        formatOp: (field, op, values, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay) => {
            let valFrom = values.first();
            let valTo = values.get(1);
            if (isForDisplay)
                return `(${field} NOT BETWEEN ${valFrom} AND ${valTo})`;
            else
                return `(${field} lt ${valFrom} or ${field} gt ${valTo})`;
        },
    }
}

const FUNCTION_OPERATORS = {
    multiselect_contains: {
        ...AntdConfig.operators.multiselect_contains,
    },
    multiselect_not_contains: {
        ...AntdConfig.operators.multiselect_not_contains
    },
    ends_with: {
        ...AntdConfig.operators.ends_with,
    },
    starts_with: {
        ...AntdConfig.operators.starts_with,
    },
    //    ...AntdConfig.operators
}

const IS_NULL_NOT_NULL = {
    is_null: {
        ...AntdConfig.operators.is_null,
        label: "is null",
        formatOp: (field, op, value, valueSrc, valueType, opDef, operatorOptions, isForDisplay) => {
            return isForDisplay ? `${field} IS NULL` : `${field} eq null`;
        },
    },
    is_not_null: {
        ...AntdConfig.operators.is_not_null,
        label: "is not null",
        formatOp: (field, op, value, valueSrc, valueType, opDef, operatorOptions, isForDisplay) => {
            return isForDisplay ? `${field} IS NOT NULL` : `${field} ne null`;
        },
    }
}

const OPERATORS = {
    //...AntdConfig.operators,
    ...BASIC_OPERATORS,
    ...BETWEEN_OPERATORS,
    ...IS_NULL_NOT_NULL
}

const BASIC_OPERATORS_LIST = Object.keys(BASIC_OPERATORS);
const FUNCTION_OPERATORS_LIST = Object.keys(FUNCTION_OPERATORS);
const BETWEEN_OPERATORS_LIST = Object.keys(BETWEEN_OPERATORS);
const IS_NULL_NOT_NULL_LIST = Object.keys(IS_NULL_NOT_NULL);

const BASIC_CONFIG = {
    ...AntdConfig,
    settings: {
        ...AntdConfig.settings,
        renderField: (props) => <FieldCascader {...props} />,
        fieldSeparator: "/",
        notLabel: "not"
    },
    conjunctions: {
        ...CONJUNCTIONS
    },
    operators: {
        ...OPERATORS
    },
    types: {
        ...AntdConfig.types,
        'group-select': {
            ...AntdConfig.types.select,
            valueSources: ['value', 'field'], //['value', 'field', 'func'],
            defaultOperator: 'equal',
            widgets: {
                'group-select': {
                    operators: IS_NULL_NOT_NULL_LIST,
                },
            },
        },
        text: {
            ...AntdConfig.types.text,
            widgets: {
                ...AntdConfig.types.text.widgets,
                // text: {
                //     operators: [
                //         "equal",
                //         "not_equal",
                //         // "starts_with",
                //         // "ends_with",
                //         ...FUNCTION_OPERATORS_LIST,
                //         ...IS_NULL_NOT_NULL_LIST
                //     ],
                //     widgetProps: {},
                //     opProps: {},
                // },
                // field: {
                //     operators: [
                //         //unary ops (like `is_empty`) will be excluded anyway, see getWidgetsForFieldOp()
                //         "equal",
                //         "not_equal",
                //         //"proximity", //can exclude if you want
                //     ],
                // }
            },
        },
        number: {
            ...AntdConfig.types.number,
            //valueSources: [], //['value', 'field', 'func'],
            widgets: {
                number: {
                    operators: [...BASIC_OPERATORS_LIST, ...BETWEEN_OPERATORS_LIST, ...IS_NULL_NOT_NULL_LIST]
                },
            },
        },
        datetime: {
            ...AntdConfig.types.datetime,
            widgets: {
                datetime: {
                    operators: [...BASIC_OPERATORS_LIST, ...BETWEEN_OPERATORS_LIST, ...IS_NULL_NOT_NULL_LIST],
                },
            },
        },
        select: {
            mainWidget: "select",
            defaultOperator: "select_equals",
            widgets: {
                select: {
                    operators: [
                        "equal",
                        "not_equal",
                        ...IS_NULL_NOT_NULL_LIST
                    ],
                },
            }
        },
        // array: {
        //     ...AntdConfig.types['!struct'],
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
        ...AntdConfig.widgets,
        'group-select': {
            ...AntdConfig.widgets.select,
        },
    },
    fieldSources: ["field", "func"],
    fields: {}
};

export { BASIC_CONFIG }; 