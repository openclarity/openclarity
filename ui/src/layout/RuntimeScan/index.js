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
                isInProgress={[PROPRESS_STATUSES.IN_PROGRESS.value, PROPRESS_STATUSES.FINALIZING.value].includes(status)}
                onStartScan={doStartScan}
                onStopScan={doStopScan}
            />
            {loading ? <Loader /> :
                <PageContainer className="scan-details-container">
                    <ProgressStep
                        title={PROPRESS_STATUSES[status].title}
                        isDone={[PROPRESS_STATUSES.DONE.value, PROPRESS_STATUSES.FINALIZING.value].includes(status)}
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
            {!loading && status === PROPRESS_STATUSES.FINALIZING.value &&
                <div className="generating-results-loader">
                    <Loader large />
                    <div>Generating results...</div>
                </div>
            }

        </div>
    )
}

export default RuntimeScan;