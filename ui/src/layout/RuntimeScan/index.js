import React from 'react';
import { isEmpty } from 'lodash';
import PageContainer from 'components/PageContainer';
import Loader from 'components/Loader';
import TopBarTitle from 'components/TopBarTitle';
import { BoldText } from 'utils/utils';
import useProgressLoaderReducer, { PROGRESS_LOADER_ACTIONS, PROPRESS_STATUSES } from './useProgressLoaderReducer';
import ScanConfiguration from './ScanConfiguration';
import ProgressStep from './ProgressStep';
import SeverityFilterAndCountersSteps from './SeverityFilterAndCountersSteps';
import TotalDisplayStep from './TotalDisplayStep';

import './runtime-scan.scss';

const RuntimeScan = () => {
    const [{loading, status, progress, scanResults, scannedNamespaces}, dispatch] = useProgressLoaderReducer();
    const doStartScan = (namespaces) => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_START_SCAN, payload: {namespaces}});
    const doStopScan = () => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_STOP_SCAN});

    const {failures, vulnerabilityPerSeverity, cisDockerBenchmarkCountPerLevel, cisDockerBenchmarkScanEnabled} = scanResults || {};
    const isInProgress = [PROPRESS_STATUSES.IN_PROGRESS.value, PROPRESS_STATUSES.FINALIZING.value].includes(status);
    
    return (
        <div className="runtime-scan-page">
            <TopBarTitle title="Runtime scan" />
            <ScanConfiguration
                isInProgress={isInProgress}
                onStartScan={doStartScan}
                onStopScan={doStopScan}
            />
            {loading ? <Loader /> :
                <PageContainer className="scan-details-container">
                    {status !== PROPRESS_STATUSES.NOT_STARTED.value &&
                        <div className="scan-details-info">
                            {!!scannedNamespaces &&
                                <React.Fragment>
                                    <BoldText>QUICK SCAN</BoldText>
                                    <span>{` with target namespaces: `}</span>
                                    <BoldText>{isEmpty(scannedNamespaces) ? "All" : scannedNamespaces.join(", ")}</BoldText>
                                </React.Fragment>
                            }
                        </div>
                    }
                    <div className="scan-details-steps">
                        <ProgressStep
                            title={PROPRESS_STATUSES[status].title}
                            isDone={[PROPRESS_STATUSES.DONE.value, PROPRESS_STATUSES.FINALIZING.value].includes(status)}
                            percent={progress}
                            scanErrors={(failures || []).map(failure => failure.message)}
                        />
                        {status === PROPRESS_STATUSES.DONE.value &&
                            <React.Fragment>
                                <TotalDisplayStep
                                    vulnerabilityPerSeverity={vulnerabilityPerSeverity}
                                    cisDockerBenchmarkCountPerLevel={cisDockerBenchmarkCountPerLevel}
                                    cisDockerBenchmarkScanEnabled={cisDockerBenchmarkScanEnabled}
                                />
                                <SeverityFilterAndCountersSteps cisDockerBenchmarkScanEnabled={cisDockerBenchmarkScanEnabled} />
                            </React.Fragment>
                        }
                    </div>
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