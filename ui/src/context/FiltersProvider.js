import { create } from './utils';

export const FILTER_TYPES = {
    APPLICATIONS: "APPLICATIONS",
    APPLICATION_RESOURCES: "APPLICATION_RESOURCES",
    PACKAGES: "PACKAGES",
    PACKAGE_RESOURCES: "PACKAGE_RESOURCES",
    VULNERABILITIES: "VULNERABILITIES"
}

const initialState = {
    ...Object.keys(FILTER_TYPES).reduce((acc, curr) => ({
        ...acc,
        [curr]: {
            tableFilters: [],
            systemFilters: {},
            selectedPageIndex: 0
        }
    }), {}),
    currentRuntimeScan: null
}

const FITLER_ACTIONS = {
    SET_TABLE_FILTERS_BY_KEY: "SET_TABLE_FILTERS_BY_KEY",
    SET_SYSTEM_FILTERS_BY_KEY: "SET_SYSTEM_FILTERS_BY_KEY",
    SET_TABLE_PAGE_BY_KEY: "SET_TABLE_PAGE_BY_KEY",
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
                    tableFilters: filterData,
                    selectedPageIndex: 0
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
                    systemFilters: filterData,
                    selectedPageIndex: 0
                }
            };
        }
        case FITLER_ACTIONS.SET_TABLE_PAGE_BY_KEY: {
            const {filterType, pageIndex} = action.payload;

            return {
                ...state,
                [filterType]: {
                    ...state[filterType],
                    selectedPageIndex: pageIndex
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
const setPage = (dispatch, {type, pageIndex}) => dispatch({type: FITLER_ACTIONS.SET_TABLE_PAGE_BY_KEY, payload: {filterType: type, pageIndex}});
const resetAllFilters = (dispatch) => dispatch({type: FITLER_ACTIONS.RESET_ALL_FILTERS});
const resetFilters = (dispatch, filterType) => dispatch({type: FITLER_ACTIONS.RESET_FILTERS_BY_KEY, payload: {filterType}});
const resetSystemFilters = (dispatch, type) => setFilters(dispatch, {type, filters: {}, isSystem: true})

export {
    FiltersProvider,
    useFilterState,
    useFilterDispatch,
    setFilters,
    setPage,
    resetAllFilters,
    resetFilters,
    resetSystemFilters
};