import { isArray, isObject } from "lodash";
import { create } from "./utils";

export const FILTER_TYPES = {
  ASSETS: "ASSETS",
  ASSET_SCANS: "ASSET_SCANS",
  SCANS: "SCANS",
  SCAN_CONFIGURATIONS: "SCAN_CONFIGURATIONS",
  FINDINGS_GENERAL: "FINDINGS_GENERAL",
  FINDINGS_VULNERABILITIES: "FINDINGS_GENERAL",
  FINDINGS_EXPLOITS: "FINDINGS_EXPLOITS",
  FINDINGS_MISCONFIGURATIONS: "FINDINGS_MISCONFIGURATIONS",
  FINDINGS_SECRETS: "FINDINGS_SECRETS",
  FINDINGS_MALWARE: "FINDINGS_MALWARE",
  FINDINGS_ROOTKITS: "FINDINGS_ROOTKITS",
  FINDINGS_PACKAGES: "FINDINGS_PACKAGES",
};

const initialState = {
  ...Object.keys(FILTER_TYPES).reduce(
    (acc, curr) => ({
      ...acc,
      [curr]: {
        tableFilters: [],
        systemFilters: {},
        customFilters: {},
        selectedPageIndex: 0,
        tableSort: {},
      },
    }),
    {},
  ),
  initialized: false,
};

const FILTER_ACTIONS = {
  SET_TABLE_FILTERS_BY_KEY: "SET_TABLE_FILTERS_BY_KEY",
  SET_SYSTEM_FILTERS_BY_KEY: "SET_SYSTEM_FILTERS_BY_KEY",
  SET_CUSTOM_FILTERS_BY_KEY: "SET_CUSTOM_FILTERS_BY_KEY",
  SET_TABLE_PAGE_BY_KEY: "SET_TABLE_PAGE_BY_KEY",
  SET_TABLE_SORT_BY_KEY: "SET_TABLE_SORT_BY_KEY",
  RESET_ALL_FILTERS: "RESET_ALL_FILTERS",
  RESET_FILTERS_BY_KEY: "RESET_FILTERS_BY_KEY",
  INITIALIZE_FILTERS: "INITIALIZE_FILTERS",
};

const reducer = (state, action) => {
  switch (action.type) {
    case FILTER_ACTIONS.SET_TABLE_FILTERS_BY_KEY: {
      const { filterType, filterData } = action.payload;

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          tableFilters: filterData,
          selectedPageIndex: 0,
        },
      };
    }
    case FILTER_ACTIONS.SET_SYSTEM_FILTERS_BY_KEY: {
      const { filterType, filterData } = action.payload;

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          tableFilters: [...initialState[filterType].tableFilters],
          systemFilters: filterData,
          selectedPageIndex: 0,
        },
      };
    }
    case FILTER_ACTIONS.SET_CUSTOM_FILTERS_BY_KEY: {
      const { filterType, filterData } = action.payload;

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          customFilters: filterData,
          selectedPageIndex: 0,
        },
      };
    }
    case FILTER_ACTIONS.SET_TABLE_PAGE_BY_KEY: {
      const { filterType, pageIndex } = action.payload;

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          selectedPageIndex: pageIndex,
        },
      };
    }
    case FILTER_ACTIONS.SET_TABLE_SORT_BY_KEY: {
      const { filterType, tableSort } = action.payload;

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          tableSort,
        },
      };
    }
    case FILTER_ACTIONS.RESET_ALL_FILTERS: {
      return Object.keys(initialState).reduce(
        (acc, curr) => ({
          ...acc,
          [curr]: {
            ...initialState[curr],
            tableSort: state[curr].tableSort,
          },
        }),
        {},
      );
    }
    case FILTER_ACTIONS.RESET_FILTERS_BY_KEY: {
      const { filterTypes } = action.payload;

      return {
        ...state,
        ...filterTypes.reduce(
          (acc, curr) => ({
            ...acc,
            [curr]: {
              ...initialState[curr],
              tableSort: state[curr].tableSort,
            },
          }),
          {},
        ),
      };
    }
    case FILTER_ACTIONS.INITIALIZE_FILTERS: {
      const {
        filterType,
        systemFilterType,
        tableFilters,
        systemFilters,
        customFilters,
      } = action.payload;

      if (
        !Object.values(FILTER_TYPES).includes(filterType) ||
        (!Object.values(FILTER_TYPES).includes(systemFilterType) &&
          !!systemFilterType) ||
        !isArray(tableFilters || {}) ||
        !isObject(systemFilters || {}) ||
        !isObject(customFilters || {})
      ) {
        return {
          ...state,
          initialized: true,
        };
      }

      return {
        ...state,
        [filterType]: {
          ...state[filterType],
          tableFilters,
          systemFilters,
          customFilters,
          selectedPageIndex: 0,
        },
        ...(!systemFilterType
          ? {}
          : {
              [systemFilterType]: {
                ...state[systemFilterType],
                systemFilters,
              },
            }),
        initialized: true,
      };
    }
    default:
      return state;
  }
};

const [FiltersProvider, useFilterState, useFilterDispatch] = create(
  reducer,
  initialState,
);

const setFilters = (
  dispatch,
  { type, filters, isSystem = false, isCustom = false },
) =>
  dispatch({
    type: isSystem
      ? FILTER_ACTIONS.SET_SYSTEM_FILTERS_BY_KEY
      : isCustom
        ? FILTER_ACTIONS.SET_CUSTOM_FILTERS_BY_KEY
        : FILTER_ACTIONS.SET_TABLE_FILTERS_BY_KEY,
    payload: { filterType: type, filterData: filters },
  });
const setPage = (dispatch, { type, pageIndex }) =>
  dispatch({
    type: FILTER_ACTIONS.SET_TABLE_PAGE_BY_KEY,
    payload: { filterType: type, pageIndex },
  });
const setSort = (dispatch, { type, tableSort }) =>
  dispatch({
    type: FILTER_ACTIONS.SET_TABLE_SORT_BY_KEY,
    payload: { filterType: type, tableSort },
  });
const resetAllFilters = (dispatch) =>
  dispatch({ type: FILTER_ACTIONS.RESET_ALL_FILTERS });
const resetFilters = (dispatch, filterTypes) =>
  dispatch({
    type: FILTER_ACTIONS.RESET_FILTERS_BY_KEY,
    payload: { filterTypes },
  });
const resetSystemFilters = (dispatch, type) =>
  setFilters(dispatch, { type, filters: {}, isSystem: true });
const initializeFilters = (dispatch, filtersData) =>
  dispatch({ type: FILTER_ACTIONS.INITIALIZE_FILTERS, payload: filtersData });

export {
  FiltersProvider,
  useFilterState,
  useFilterDispatch,
  setFilters,
  setPage,
  setSort,
  resetAllFilters,
  resetFilters,
  resetSystemFilters,
  initializeFilters,
};
