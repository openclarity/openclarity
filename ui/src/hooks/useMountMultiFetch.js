import { useEffect } from 'react';
import { isEqual } from 'lodash';
import { usePrevious } from 'hooks';
import useMultiFetch from './useMultiFetch';

const useMountMultiFetch = (urlsData) => {
    const [state, fetchData] = useMultiFetch({initialLoading: true});
    
    const prevUrlData = usePrevious(urlsData);

    useEffect(() => {
        if (!isEqual(urlsData, prevUrlData)) {
            fetchData(urlsData);
        }
    }, [urlsData, prevUrlData, fetchData]);

    return [state, () => fetchData(urlsData)];
}

export default useMountMultiFetch;