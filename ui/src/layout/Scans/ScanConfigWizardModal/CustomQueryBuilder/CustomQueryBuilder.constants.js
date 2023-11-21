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

const FUNCTION_TEXT_OPERATORS = {
    ends_with: {
        ...AntdConfig.operators.ends_with,
        label: "ends with",
        labelForFormat: "Ends with",
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            if (isForDisplay)
                return `(${field} ENDS WITH ${value})`;
            else
                return `endswith(${field}, ${value})`;
        },
    },
    starts_with: {
        ...AntdConfig.operators.starts_with,
        label: "starts with",
        labelForFormat: "Starts with",
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            if (isForDisplay)
                return `(${field} STARTS WITH ${value})`;
            else
                return `startswith(${field}, ${value})`;
        },
    },
}

/*
const LAMBDA_FUNCTION_OPERATORS = {
    any: {
        label: "any",
        labelForFormat: "Any",
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            if (isForDisplay)
                return `(${field} HAS ANY ${value})`;
            else
                return `(${field}/any(o:o/${value} eq ${"sg"})`;
        },
    },
    all: {
        label: "all",
        labelForFormat: "All",
        formatOp: (field, op, value, valueSrcs, valueTypes, opDef, operatorOptions, isForDisplay, fieldDef) => {
            if (isForDisplay)
                return `(${field} ALL MATCH ${value})`;
            else
                return `(${field}/all(o:o/${value} eq ${"sg"})`;
        },
        //valueTypes: ['number']
    }
}
*/

const IS_NULL_NOT_NULL_OPERATORS = {
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
    ...BASIC_OPERATORS,
    ...BETWEEN_OPERATORS,
    ...IS_NULL_NOT_NULL_OPERATORS,
    ...FUNCTION_TEXT_OPERATORS,
    //...LAMBDA_FUNCTION_OPERATORS
}

const BASIC_OPERATORS_LIST = Object.keys(BASIC_OPERATORS);
const FUNCTION_TEXT_OPERATORS_LIST = Object.keys(FUNCTION_TEXT_OPERATORS);
const BETWEEN_OPERATORS_LIST = Object.keys(BETWEEN_OPERATORS);
const IS_NULL_NOT_NULL_LIST = Object.keys(IS_NULL_NOT_NULL_OPERATORS);
//const LAMBDA_FUNCTION_OPERATORS_LIST = Object.keys(LAMBDA_FUNCTION_OPERATORS)

const TYPES = {
    ...AntdConfig.types,
    'group-select': {
        ...AntdConfig.types.select,
        valueSources: ['value', 'field'],
        defaultOperator: 'is_null',
        widgets: {
            'group-select': {
                operators: [
                    ...IS_NULL_NOT_NULL_LIST,
                ]

            },
        },
    },
    'array-select': {
        ...AntdConfig.types.select,
        valueSources: ['value', 'field', 'func'],
        defaultOperator: 'is_null',
        widgets: {
            'group-select': {
                operators: [
                    ...IS_NULL_NOT_NULL_LIST,
                    //...LAMBDA_FUNCTION_OPERATORS_LIST
                ]
            },
        },
    },
    text: {
        ...AntdConfig.types.text,
        valueSources: ['value', 'func', 'field'],
        widgets: {
            ...AntdConfig.types.text.widgets,
            text: {
                ...AntdConfig.types.text.widgets.text,
                operators: [
                    ...BASIC_OPERATORS_LIST,
                    ...FUNCTION_TEXT_OPERATORS_LIST,
                    ...IS_NULL_NOT_NULL_LIST,
                ],
            },
            number: {
                ...AntdConfig.types.text.widgets.number,
                operators: [
                    ...IS_NULL_NOT_NULL_LIST
                ]
            },
        },
    },
    number: {
        ...AntdConfig.types.number,
        widgets: {
            number: {
                operators: [
                    ...BASIC_OPERATORS_LIST,
                    ...BETWEEN_OPERATORS_LIST,
                    ...IS_NULL_NOT_NULL_LIST
                ]
            },
        },
    },
    datetime: {
        ...AntdConfig.types.datetime,
        widgets: {
            datetime: {
                operators: [
                    ...BASIC_OPERATORS_LIST,
                    ...BETWEEN_OPERATORS_LIST,
                    ...IS_NULL_NOT_NULL_LIST
                ],
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
}

const WIDGETS = {
    ...AntdConfig.widgets,
    'group-select': {
        ...AntdConfig.widgets.select,
    },
    datetime: {
        ...AntdConfig.widgets.datetime,
        timeFormat: "HH:mm:ss",
        valueFormat: `YYYY-MM-DDTHH:mm:ss`, // This is the locale datetime
        formatValue: function (val, fieldDef, wgtDef, isForDisplay) {
            const dateVal = this.utils.moment(val, wgtDef.valueFormat);
            return isForDisplay ? this.utils.moment(dateVal).format('LLLL') : JSON.stringify(this.utils.moment.utc(dateVal).format());
        },

    },
}

const BASIC_CONFIG = {
    ...AntdConfig,

    settings: {
        ...AntdConfig.settings,
        renderField: (props) => {
            props.items = props.items.sort((a, b) => {
                if (a.key < b.key) {
                    return -1;
                } else if (a.key > b.key) {
                    return 1;
                } else {
                    return 0;
                }
            })

            return (<FieldCascader {...props} customProps={{ ...props.customProps, changeOnSelect: true }} />)
        }
        ,
        canRegroup: false,
        locale: {
            moment: navigator?.languages?.length
                ? navigator.languages[0]
                : navigator?.language
        },
        fieldSeparator: "/",
        fieldSources: ["field", "func"],
        valueSourcesInfo: {
            value: {
                label: "Value"
            },
            field: {
                label: "Field",
                widget: "field",
            },
            func: {
                label: "Function",
                widget: "func",
            }
        },
        notLabel: "not"
    },
    conjunctions: {
        ...CONJUNCTIONS
    },
    operators: {
        ...OPERATORS
    },
    types: {
        ...TYPES
    },
    widgets: {
        ...WIDGETS
    },
    funcs: {
        length: {
            label: 'length',
            /*
            formatFunc: (args, isForDisplay) => {
                console.log('isForDisplay:', isForDisplay);
                console.log('args:', args);

            },
            */
            returnType: 'text',
            args: {
                text: {
                    type: 'text',
                    valueSources: ['value', 'field'],
                },
            }
        }
    },
    fields: {}
};

export { BASIC_CONFIG }; 