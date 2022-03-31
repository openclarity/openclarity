import { create } from './utils';

export const FILTERR_TYPES = {
    APPLICATIONS: "APPLICATIONS",
    APPLICATION_RESOURCES: "APPLICATION_RESOURCES",
    PACKAGES: "PACKAGES",
    PACKAGE_RESOURCES: "PACKAGE_RESOURCES",
    VULNERABILITIES: "VULNERABILITIES"
}

const initialState = {
    [FILTERR_TYPES.APPLICATIONS]: {
        tableFilters: [],
        systemFilters: {}
    },
    [FILTERR_TYPES.APPLICATION_RESOURCES]: {
        tableFilters: [],
        systemFilters: {}
    },
    [FILTERR_TYPES.PACKAGES]: {
        tableFilters: [],
        systemFilters: {}
    },
    [FILTERR_TYPES.PACKAGE_RESOURCES]: {
        tableFilters: []
    },
    [FILTERR_TYPES.VULNERABILITIES]: {
        tableFilters: [],
        systemFilters: {}
    },
    currentRuntimeScan: null
};

const FITLER_ACTIONS = {
    SET_TABLE_FILTERS_BY_KEY: "SET_TABLE_FILTERS_BY_KEY",
    SET_SYSTEM_FILTERS_BY_KEY: "SET_SYSTEM_FILTERS_BY_KEY",
    RESET_ALL_FILTERS: "RESET_ALL_FILTERS",
    RESET_FILTERS_BY_KEY: "RESET_FILTERS_BY_KEY"
}

const reducer = (state, action) => {
    switch (action.type) {
        case FITLER_ACTIONS.SET_TABLE_FILTERS_BY_KEY: {
            const {filterType, filterData} = action.payload;

            return {
                ...state,
                [filterType]: {
                    ...state[filterType],
                    tableFilters: filterData
                }
            };
        }
        case FITLER_ACTIONS.SET_SYSTEM_FILTERS_BY_KEY: {
            const {filterType, filterData} = action.payload;
            
            return {
                ...state,
                [filterType]: {
                    ...state[filterType],
                    tableFilters: [...initialState[filterType].tableFilters],
                    systemFilters: filterData
                }
            };
        }
        case FITLER_ACTIONS.RESET_ALL_FILTERS: {
            return {
                ...state,
                ...initialState
            };
        }
        case FITLER_ACTIONS.RESET_FILTERS_BY_KEY: {
            const {filterType} = action.payload;
            
            return {
                ...state,
                [filterType]: {
                    ...initialState[filterType]
                }
            };
        }
        default:
            return state;
    }
}

const [FiltersProvider, useFilterState, useFilterDispatch] = create(reducer, initialState);

const setFilters = (dispatch, {type, filters, isSystem=false}) => dispatch({
    type: isSystem ? FITLER_ACTIONS.SET_SYSTEM_FILTERS_BY_KEY : FITLER_ACTIONS.SET_TABLE_FILTERS_BY_KEY,
    payload: {filterType: type, filterData: filters}
});
const resetAllFilters = (dispatch) => dispatch({type: FITLER_ACTIONS.RESET_ALL_FILTERS});
const resetFilters = (dispatch, filterType) => dispatch({type: FITLER_ACTIONS.RESET_FILTERS_BY_KEY, payload: {filterType}});
const resetSystemFilters = (dispatch, type) => setFilters(dispatch, {type, filters: {}, isSystem: true})

export {
    FiltersProvider,
    useFilterState,
    useFilterDispatch,
    setFilters,
    resetAllFilters,
    resetFilters,
    resetSystemFilters
};