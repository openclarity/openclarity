import { BasicConfig } from "@react-awesome-query-builder/ui";
import { AntdWidgets } from '@react-awesome-query-builder/antd';

const { FieldCascader } = AntdWidgets;

const BASIC_OPERATORS = {
    equal: {
        ...BasicConfig.operators.equal,
        label: 'equals',
        labelForFormat: 'eq',
    },
    not_equal: {
        ...BasicConfig.operators.not_equal,
        label: 'does not equal',
        labelForFormat: 'ne',
    },
    greater: {
        ...BasicConfig.operators.greater,
        label: 'greater than',
        labelForFormat: 'gt',
    },
    greater_or_equal: {
        ...BasicConfig.operators.greater_or_equal,
        label: 'greater than or equal',
        labelForFormat: 'ge',
    },
    less: {
        ...BasicConfig.operators.less,
        label: 'less than',
        labelForFormat: 'lt',
    },
    less_or_equal: {
        ...BasicConfig.operators.less_or_equal,
        label: 'less than or equal',
        labelForFormat: 'le',
    },
}

const BETWEEN_OPERATORS = {
    between: {
        ...BasicConfig.operators.between,
        label: "in between",
    },
    not_between: {
        ...BasicConfig.operators.not_between,
        label: "not in between"
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

const OPERATORS = {
    ...BasicConfig.operators,
    ...BASIC_OPERATORS,
    ...BETWEEN_OPERATORS
}

const BASIC_OPERATORS_LIST = Object.keys(BASIC_OPERATORS);
const FUNCTION_OPERATORS_LIST = Object.keys(FUNCTION_OPERATORS);
const BETWEEN_OPERATORS_LIST = Object.keys(BETWEEN_OPERATORS);
console.log('BETWEEN_OPERATORS:', BETWEEN_OPERATORS);
console.log('BETWEEN_OPERATORS_LIST:', BETWEEN_OPERATORS_LIST);

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
            formatConj: (children, _conj, not) => ((not ? 'not ' : '') + '(' + children.join(' || ') + ')'),
        },
        OR: {
            ...BasicConfig.conjunctions.OR,
            label: 'or',
            formatConj: (children, _conj, not) => ((not ? 'not ' : '') + '(' + children.join(' || ') + ')'),
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
                    operators: ['equal', 'not_equal'],
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

console.log('BASIC_CONFIG1:', BASIC_CONFIG);
export { BASIC_CONFIG }; 