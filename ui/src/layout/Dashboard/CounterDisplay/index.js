import React from 'react';
import { useFetch } from 'hooks';
import { formatNumber } from 'utils/utils';

import './counter-display.scss';

const CounterDisplay = ({url, title, background}) => {
    const [{data, error, loading}] = useFetch(url, {queryParams: {"$count": true, "$top": 1, "$select": "id"}});
    
    return (
        <div className="dashboard-counter" style={{background}}>
            {loading || error ? "" : 
                <div className="dashboard-counter-content"><div className="dashboard-counter-count">{formatNumber(data.count)}</div>{` ${title}`}</div>
            }
        </div>
    )
}

export default CounterDisplay;