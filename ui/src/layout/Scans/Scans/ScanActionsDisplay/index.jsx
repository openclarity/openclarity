import React, { useEffect } from 'react';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';
import { SCAN_STATES } from 'components/ScanProgressBar';
import { APIS } from 'utils/systemConsts';

import './scan-actions-display.scss';

const ScanActionsDisplay = ({data, onUpdate}) => {
    const {id, status: { state }} = data;

    const [{loading, error}, fetchScan] = useFetch(APIS.SCANS, {loadOnMount: false});
    const prevLoading = usePrevious(loading);

    useEffect(() => {
        if (prevLoading && !loading && !error && !!onUpdate) {
            onUpdate();
        }
    }, [loading, prevLoading, error, onUpdate])

    if ([SCAN_STATES.Done.state, SCAN_STATES.Failed.state, SCAN_STATES.Aborted.state].includes(state) || loading) {
        return null;
    }

    return (
        <div className="scan-actions-display">
            <IconWithTooltip
                tooltipId={`${id}-stop`}
                tooltipText="Stop scan"
                name={ICON_NAMES.STOP}
                onClick={event => {
                    event.stopPropagation();
                    event.preventDefault();

                    fetchScan({
                        method: FETCH_METHODS.PATCH,
                        submitData: {
                          status: {
                            state: SCAN_STATES.Aborted.state,
                            reason: "Cancellation",
                            message: "Scan has been aborted",
                            lastTransitionTime: new Date().toISOString(),
                          },
                        },
                        formatUrl: url => `${url}/${id}`
                    });
                }}
            />
        </div>
    );
}

export default ScanActionsDisplay;
