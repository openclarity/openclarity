import { useReducer, useEffect, useRef, useCallback } from 'react';
import { isUndefined, isNull, isEmpty } from 'lodash';
import { useNotificationDispatch, showNotification } from 'context/NotificationProvider';
import { NOTIFICATION_TYPES } from 'components/Notification';

export const FETCH_METHODS = {
    GET: "GET",
    POST: "POST",
    PUT: "PUT",
    DELETE: "DELETE",
    PATCH: "PATCH"
}

const FETCH_ACTIONS = {
    LOADING_DATA: "LOADING_DATA",
    LOAD_DATA_SUCCESS: "LOAD_DATA_SUCCESS",
    LOAD_DATA_ERROR: "LOAD_DATA_ERROR",
    UPDATE_FETCH_PARAMS: "UPDATE_FETCH_PARAMS"
}

const queryString = (params) => Object.keys(params).map((key) => {
    return encodeURIComponent(key) + '=' + encodeURIComponent(params[key])
}).join('&');

export const formatFetchUrl = ({ url, queryParams, formatUrl, urlPrefix, isAbsoluteUrl }) => {

    if (isAbsoluteUrl) {
        return isEmpty(queryParams) ? url : `${url}?${queryString(queryParams)}`;
    }

    const formattedUrl = !!formatUrl ? formatUrl(url) : url;
    const formattedPrefix = !!urlPrefix ? `${urlPrefix}/api` : "api";

    return isEmpty(queryParams) ? `/${formattedPrefix}/${formattedUrl}` : `/${formattedPrefix}/${formattedUrl}?${queryString(queryParams)}`;
}

export const formatFetchOptions = ({ method, stringifiedSubmitData }) => {
    const options = {
        credentials: 'include',
        method
    };

    if ([FETCH_METHODS.POST, FETCH_METHODS.PUT, FETCH_METHODS.PATCH].includes(method)) {
        options.headers = { 'content-type': 'application/json' };
        options.body = stringifiedSubmitData;
    }

    return options;
}

const getErrorMessage = (method) => `An error occurred when trying to ${method === FETCH_METHODS.GET ? "fetch" : "submit"} data`;

function reducer(state, action) {
    switch (action.type) {
        case FETCH_ACTIONS.LOADING_DATA:
            return { ...state, loading: true, error: null, loadData: false };
        case FETCH_ACTIONS.LOAD_DATA_SUCCESS:
            return { ...state, loading: false, data: action.payload, loadData: false };
        case FETCH_ACTIONS.LOAD_DATA_ERROR:
            return { ...state, loading: false, error: action.payload, loadData: false, data: null };
        case FETCH_ACTIONS.UPDATE_FETCH_PARAMS:
            const { queryParams, method = FETCH_METHODS.GET, submitData, formatUrl, urlPrefix, isAbsoluteUrl } = action.payload || {};

            return {
                ...state,
                url: formatFetchUrl({ url: state.baseUrl, queryParams, formatUrl, urlPrefix, isAbsoluteUrl }),
                method: method.toUpperCase(),
                submitData: !!submitData ? JSON.stringify(submitData) : null,
                loadData: true,
                data: undefined
            };
        default:
            return { ...state };
    }
}

function useFetch(baseUrl, options) {
    const { queryParams, method: initialMethod, submitData: inititalSubmitData, formatUrl, loadOnMount = true, urlPrefix, isAbsoluteUrl = false } = options || {};

    const [state, dispatch] = useReducer(reducer, {
        loading: false,
        error: null,
        data: loadOnMount ? undefined : null,
        baseUrl,
        url: formatFetchUrl({ url: baseUrl, queryParams, formatUrl, urlPrefix, isAbsoluteUrl }),
        method: !!initialMethod ? initialMethod.toUpperCase() : FETCH_METHODS.GET,
        submitData: !!inititalSubmitData ? JSON.stringify(inititalSubmitData) : null,
        loadData: loadOnMount || false
    });

    const mounted = useRef(true);

    useEffect(() => {
        return function cleanup() {
            mounted.current = false;
        };
    }, []);

    const notificationDispatch = useNotificationDispatch();

    const { url, method, submitData, loadData, data, error, loading } = state;

    const doFetch = useCallback(async () => {
        const options = formatFetchOptions({ method, stringifiedSubmitData: submitData });

        dispatch({ type: FETCH_ACTIONS.LOADING_DATA });

        let isError = false;
        const showErrorMessage = () => showNotification(notificationDispatch, { message: getErrorMessage(method), type: NOTIFICATION_TYPES.ERROR });

        fetch(url, options)
            .then(response => {
                isError = !response.ok;

                return response;
            })
            .then(response => response.status === 204 ? {} : response.json())
            .then(data => {
                if (!mounted.current) {
                    return;
                }

                if (isError) {
                    dispatch({ type: FETCH_ACTIONS.LOAD_DATA_ERROR, payload: data });

                    showErrorMessage();

                    return;
                }

                dispatch({ type: FETCH_ACTIONS.LOAD_DATA_SUCCESS, payload: data });
            })
            .catch(error => {
                if (!mounted.current) {
                    return;
                }

                showErrorMessage();

                dispatch({ type: FETCH_ACTIONS.LOAD_DATA_ERROR, payload: error });
            });
    }, [url, method, submitData, notificationDispatch]);

    useEffect(() => {
        if (!mounted.current) {
            return;
        }

        if (!loadData) {
            return;
        }

        doFetch();
    }, [doFetch, loadOnMount, loadData]);

    const fetchData = useCallback(fetchParams => dispatch({ type: FETCH_ACTIONS.UPDATE_FETCH_PARAMS, payload: fetchParams }), []);

    return [{ data, error, loading: loading || (isUndefined(data) && isNull(error)) }, fetchData];
}

export default useFetch;