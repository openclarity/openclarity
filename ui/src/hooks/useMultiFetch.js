import { useReducer, useEffect, useRef } from 'react';
import { useNotificationDispatch, showNotification } from 'context/NotificationProvider'; 
import { NOTIFICATION_TYPES } from 'components/Notification';
import { formatFetchUrl, formatFetchOptions, FETCH_METHODS } from './useFetch';

const FETCH_ACTIONS = {
    LOADING_DATA: "LOADING_DATA",
    LOAD_DATA_SUCCESS: "LOAD_DATA_SUCCESS",
    LOAD_DATA_ERROR: "LOAD_DATA_ERROR"
}

function reducer(state, action) {
    switch (action.type) {
        case FETCH_ACTIONS.LOADING_DATA:
            return {...state, loading: true, error: null};
        case FETCH_ACTIONS.LOAD_DATA_SUCCESS:
            return {...state, loading: false, data: action.payload};
        case FETCH_ACTIONS.LOAD_DATA_ERROR:
            return {...state, loading: false, error: action.payload};
        default:
            return {...state};
    }
}

function useMultiFetch(options) {
    const notificationDispatch = useNotificationDispatch();

    const {initialLoading=false} = options || {};
    const [state, dispatch] = useReducer(reducer, {
        loading: initialLoading,
        error: null,
        data: null
    });

    const mounted = useRef(true);

    useEffect(() => {
        return function cleanup() {
            mounted.current = false;
        };
    }, []);

    const fetchData = async (urlsData) => {
        dispatch({type: FETCH_ACTIONS.LOADING_DATA});

        try {
            const response = await Promise.all(
                urlsData.map(urlData => {
                    const {url, queryParams, data, method=FETCH_METHODS.GET, formatUrl} = urlData;
                    
                    const options = formatFetchOptions({
                        method: method.toUpperCase(),
                        stringifiedSubmitData: JSON.stringify(data)
                    });

                    return fetch(formatFetchUrl({url, queryParams, formatUrl}), options)
                        .then(response => {
                            if (!response.ok) {
                                throw Error(response.statusText);
                            }
    
                            return response;
                        })
                        .then(response => response.json())
                })
            );
            
            const data = response.map((item, index) => ({key: urlsData[index].key, data: item})).reduce((accumulator, curr) => {
                accumulator[curr.key] = curr.data; 
                return accumulator
            }, {});
            
            if (!mounted.current) {
                return;
            }
            
            dispatch({type: FETCH_ACTIONS.LOAD_DATA_SUCCESS, payload: data});
        } catch (error) {
            if (!mounted.current) {
                return;
            }
            
            showNotification(notificationDispatch, {message: "An error occurred when trying to fetch data", type: NOTIFICATION_TYPES.ERROR});

            dispatch({type: FETCH_ACTIONS.LOAD_DATA_ERROR, payload: error});
        }
    };

    return [state, fetchData];
}

export default useMultiFetch;