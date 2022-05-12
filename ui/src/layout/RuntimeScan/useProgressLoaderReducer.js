import { useReducer, useEffect, useRef } from 'react';
import { usePrevious, useFetch, FETCH_METHODS } from 'hooks';

export const PROPRESS_STATUSES = {
    NOT_STARTED: {value: "NOT_STARTED", title: "Click above to start"},
    IN_PROGRESS: {value: "IN_PROGRESS", title: "Scanning..."},
    FINALIZING: {value: "FINALIZING", title: "Scan complete"},
    DONE: {value: "DONE", title: "Scan complete"}
}

const initialState = {
    isLoading: true,
    isLoadingError: false,
    loadData: false,
    progress: 0,
    status: PROPRESS_STATUSES.NOT_STARTED.value,
    namespacesToScan: null,
    doAbort: false,
    scanResults: null,
    scannedNamespaces: null,
    scanType: null,
    startTime: null
};

export const PROGRESS_LOADER_ACTIONS = {
    ERROR_LOADIND_STATUS: "ERROR_LOADIND_STATUS",
    STATUS_DATA_LOADED: "STATUS_DATA_LOADED",
    DO_STOP_SCAN: "DO_STOP_SCAN",
    DO_START_SCAN: "DO_START_SCAN",
    SCAN_STOPPED: "SCAN_STOPPED",
    RESULTS_DATA_LOADED: "RESULTS_DATA_LOADED"
}

export const RUNTIME_SCAN_URL = "runtime/scan";

const reducer = (state, action) => {
    switch (action.type) {
        case PROGRESS_LOADER_ACTIONS.STATUS_DATA_LOADED: {
            const {status, progress, scannedNamespaces, scanType, startTime} = action.payload;

            return {
                ...state,
                isLoading: false,
                status,
                progress,
                namespacesToScan: null,
                doAbort: false,
                doLoadStatus: false,
                scanResults: null,
                scannedNamespaces,
                scanType,
                startTime
            };
        }
        case PROGRESS_LOADER_ACTIONS.ERROR_LOADIND_STATUS: {
            return {
                ...state,
                isLoading: false,
                isLoadingError: true,
                scanResults: null
            };
        }
        case PROGRESS_LOADER_ACTIONS.DO_START_SCAN: {
            const {namespaces} = action.payload;

            return {
                ...state,
                isLoading: true,
                status: PROPRESS_STATUSES.IN_PROGRESS.value,
                progress: 0,
                namespacesToScan: namespaces,
                scanResults: null
            };
        }
        case PROGRESS_LOADER_ACTIONS.DO_STOP_SCAN: {
            return {
                ...state,
                isLoading: true,
                status: PROPRESS_STATUSES.NOT_STARTED.value,
                progress: 0,
                doAbort: true
            };
        }
        case PROGRESS_LOADER_ACTIONS.SCAN_STOPPED: {
            return {
                ...state,
                isLoading: false,
                doAbort: false
            };
        }
        case PROGRESS_LOADER_ACTIONS.RESULTS_DATA_LOADED: {
            return {
                ...state,
                scanResults: action.payload
            };
        }
        default:
            return state;
    }
}

function useProgressLoaderReducer() {
    const [{isLoading, isLoadingError, loadData, status, progress, doAbort, namespacesToScan, scanResults, scannedNamespaces, scanType, startTime}, dispatch] =
        useReducer(reducer, {...initialState});
    const prevDoAbort = usePrevious(doAbort);
    const prevNamespacesToScan = usePrevious(namespacesToScan);
    const prevStatus = usePrevious(status);
    
    const [{loading, data, error}, fetchStatus] = useFetch(`${RUNTIME_SCAN_URL}/progress`);
    const prevLoading = usePrevious(loading);

    const [{loading: stopping}, stopScan] = useFetch(`${RUNTIME_SCAN_URL}/stop`, {loadOnMount: false});
    const prevStopping = usePrevious(stopping);

    const [{loading: starting, error: startError}, startScan] = useFetch(`${RUNTIME_SCAN_URL}/start`, {loadOnMount: false});
    const prevStarting = usePrevious(starting);

    const [{loading: loadingResults, data: results, error: resultsError}, fetchResults] = useFetch(`${RUNTIME_SCAN_URL}/results`, {loadOnMount: false});
    const prevLoadingResults = usePrevious(loadingResults);
    
    const fetcherRef = useRef(null);

    useEffect(() => {
        return function cleanup() {
            if (fetcherRef.current) {
                clearTimeout(fetcherRef.current);
            }
        };
    }, []);
    
    useEffect(() => {
        if (prevLoading && !loading) {
            if (!!error) {
                dispatch({type: PROGRESS_LOADER_ACTIONS.ERROR_LOADIND_STATUS});
            } else {
                const {scanned, status, scannedNamespaces, scanType, startTime} = data;

                dispatch({type: PROGRESS_LOADER_ACTIONS.STATUS_DATA_LOADED, payload: {progress: scanned, status, scannedNamespaces, scanType, startTime}});
                
                if ([PROPRESS_STATUSES.IN_PROGRESS.value, PROPRESS_STATUSES.FINALIZING.value].includes(status)) {
                    fetcherRef.current = setTimeout(() => fetchStatus(), 3000);
                }
            }
        }
    }, [prevLoading, loading, data, error, fetchStatus]);

    useEffect(() => {
        if (!prevDoAbort && doAbort) {
            clearTimeout(fetcherRef.current);
            stopScan({method: FETCH_METHODS.PUT});
        }
    }, [prevDoAbort, doAbort, stopScan]);

    useEffect(() => {
        if (!prevNamespacesToScan && !!namespacesToScan) {
            startScan({method: FETCH_METHODS.PUT, submitData: {namespaces: namespacesToScan}});
        }
    }, [prevNamespacesToScan, namespacesToScan, fetchStatus, startScan]);

    useEffect(() => {
        if (prevStarting && !starting && !startError) {
            fetchStatus();
        }
    }, [prevStarting, starting, startError, fetchStatus]);

    useEffect(() => {
        if (prevStopping && !stopping) {
            dispatch({type: PROGRESS_LOADER_ACTIONS.SCAN_STOPPED});
        }
    }, [prevStopping, stopping, fetchStatus]);

    useEffect(() => {
        if (status !== prevStatus && status === PROPRESS_STATUSES.DONE.value) {
            fetchResults();
        }
    }, [prevStatus, status, fetchResults]);

    useEffect(() => {
        if (!loadData) {
            return;
        }
        
        fetchStatus();
    }, [fetchStatus, loadData]);

    useEffect(() => {
        if (prevLoadingResults && !loadingResults && !resultsError) {
            dispatch({type: PROGRESS_LOADER_ACTIONS.RESULTS_DATA_LOADED, payload: results});
        }
    }, [prevLoadingResults, loadingResults, resultsError, results]);

    return [{loading: isLoading, isLoadingError, status, progress, scanResults, scannedNamespaces, scanType, startTime}, dispatch];
}

export default useProgressLoaderReducer;