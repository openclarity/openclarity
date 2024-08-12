import { useReducer, useEffect, useRef, useCallback } from 'react';
import { useNotificationDispatch, showNotification } from 'context/NotificationProvider'; 
import { NOTIFICATION_TYPES } from 'components/Notification';
import { formatFetchOptions, FETCH_METHODS } from './useFetch';

const DELETE_ACTIONS = {
    DELETING: "DELETING",
    DELETE_SUCCESS: "DELETE_SUCCESS",
    DELETE_ERROR: "DELETE_ERROR",
    UPDATE_DELETE_PARAMS: "UPDATE_DELETE_PARAMS",
    DELETE_GENERAL_ERROR_MESSAGE: "DELETE_GENERAL_ERROR_MESSAGE"
}

function reducer(state, action) {
    switch (action.type) {
        case DELETE_ACTIONS.DELETING:
            return {...state, deleting: true, error: null, callDelete: false};
        case DELETE_ACTIONS.DELETE_SUCCESS:
            return {...state, deleting: false, callDelete: false};
        case DELETE_ACTIONS.DELETE_ERROR:
            return {...state, deleting: false, error: action.payload, callDelete: false};
        case DELETE_ACTIONS.UPDATE_DELETE_PARAMS:
            return {
                ...state,
                callDelete: true,
                formattedUrl: `/api/${state.baseUrl}/${action.payload}${state.urlSuffix}`
            };
        default:
            return {...state};
    }
}
const useDelete = (url, options) => {
    const notificationDispatch = useNotificationDispatch();

    const {urlSuffix, showServerError} = options || {};
    const [state, dispatch] = useReducer(reducer, {
        deleting: false,
        error: null,
        baseUrl: url,
        formattedUrl: null,
        urlSuffix: urlSuffix || "",
        callDelete: false
    });

    const mounted = useRef(true);

    useEffect(() => {
        return function cleanup() {
            mounted.current = false;
        };
    }, []);

    const {error, deleting, formattedUrl, callDelete} = state;
    
    const doDelete = useCallback(async () => {
        const options = formatFetchOptions({method: FETCH_METHODS.DELETE});

        dispatch({type: DELETE_ACTIONS.DELETING});

        let isError = false;
        const showErrorMessage = (message) => showNotification(notificationDispatch, {message: showServerError && !!message ? message : DELETE_ACTIONS.DELETE_GENERAL_ERROR_MESSAGE, type: NOTIFICATION_TYPES.ERROR});

        fetch(formattedUrl, options)
            .then(response => {
                isError = !response.ok;

                return response;
            })
            .then(response => response.status === 204 ? {} : response.json())
            .then(data => {
                if (isError) {
                    dispatch({type: DELETE_ACTIONS.DELETE_ERROR, payload: data});

                    showErrorMessage(data.message);

                    return;
                }

                if (!mounted.current) {
                    return;
                }

                dispatch({type: DELETE_ACTIONS.DELETE_SUCCESS, payload: data});
            })
            .catch(error => {
                if (!mounted.current) {
                    return;
                }

                showErrorMessage();

                dispatch({type: DELETE_ACTIONS.DELETE_ERROR, payload: error});
            });
    }, [formattedUrl, showServerError, notificationDispatch]);

    useEffect(() => {
        if (!formattedUrl || !callDelete) {
            return;
        }

        doDelete();
    }, [doDelete, formattedUrl, callDelete]);

    const deleteData = deleteId => dispatch({type: DELETE_ACTIONS.UPDATE_DELETE_PARAMS, payload: deleteId});
    
    return [{error, deleting}, deleteData];
}

export default useDelete;