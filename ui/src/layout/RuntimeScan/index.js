import React from 'react';
import PageContainer from 'components/PageContainer';
import VulnerabilitiesSummaryDisplay from 'components/VulnerabilitiesSummaryDisplay';
import Loader from 'components/Loader';
import TopBarTitle from 'components/TopBarTitle';
import useProgressLoaderReducer, { PROGRESS_LOADER_ACTIONS, PROPRESS_STATUSES } from './useProgressLoaderReducer';
import StepDisplay from './StepDisplay';
import ScanConfiguration from './ScanConfiguration';
import ProgressStep from './ProgressStep';
import SeverityCountersStep from './SeverityCountersStep';

import './runtime-scan.scss';

const RuntimeScan = () => {
    const [{loading, status, progress, scanResults}, dispatch] = useProgressLoaderReducer();
    const doStartScan = (namespaces) => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_START_SCAN, payload: {namespaces}});
    const doStopScan = () => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_STOP_SCAN});

    const {failures, vulnerabilityPerSeverity} = scanResults || {};
    
    return (
        <div className="runtime-scan-page">
            <TopBarTitle title="Runtime scan" />
            <ScanConfiguration
                isInProgress={status === PROPRESS_STATUSES.IN_PROGRESS.value}
                onStartScan={doStartScan}
                onStopScan={doStopScan}
            />
            {loading ? <Loader /> :
                <PageContainer className="scan-details-container">
                    <ProgressStep
                        title={PROPRESS_STATUSES[status].title}
                        isDone={status === PROPRESS_STATUSES.DONE.value}
                        percent={progress}
                        scanErrors={(failures || []).map(failure => failure.message)}
                    />
                    {status === PROPRESS_STATUSES.DONE.value &&
                        <React.Fragment>
                            <StepDisplay step="2"  title="Total vulnerabilities:">
                                <VulnerabilitiesSummaryDisplay id="runtime-scan" vulnerabilities={vulnerabilityPerSeverity || []} />
                            </StepDisplay>
                            <SeverityCountersStep />
                        </React.Fragment>
                    }
                </PageContainer>
            }
        </div>
    )
}

export default RuntimeScan;