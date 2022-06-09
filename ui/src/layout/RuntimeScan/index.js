import React from 'react';
import { isEmpty } from 'lodash';
import PageContainer from 'components/PageContainer';
import Loader from 'components/Loader';
import TopBarTitle from 'components/TopBarTitle';
import Icon, { ICON_NAMES } from 'components/Icon';
import { BoldText, formatDate } from 'utils/utils';
import useProgressLoaderReducer, { PROGRESS_LOADER_ACTIONS, PROPRESS_STATUSES } from './useProgressLoaderReducer';
import ScanConfiguration from './ScanConfiguration';
import ProgressStep from './ProgressStep';
import SeverityFilterAndCountersSteps from './SeverityFilterAndCountersSteps';
import TotalDisplayStep from './TotalDisplayStep';

import './runtime-scan.scss';

const SCAN_TYPES = {
    QUICK: "ON-DEMAND",
    SCHEDULE: "SCHEDULES"
}

const RuntimeScan = () => {
    const [{loading, status, progress, scanResults, scannedNamespaces, scanType, startTime}, dispatch] = useProgressLoaderReducer();
    const doStartScan = (namespaces) => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_START_SCAN, payload: {namespaces}});
    const doStopScan = () => dispatch({type: PROGRESS_LOADER_ACTIONS.DO_STOP_SCAN});

    const {failures, vulnerabilityPerSeverity, cisDockerBenchmarkCountPerLevel, cisDockerBenchmarkScanEnabled, endTime} = scanResults || {};
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
                            {!!scannedNamespaces && !!scanType &&
                                <React.Fragment>
                                    <div className="genreral-scan-details">
                                        <BoldText>{`${SCAN_TYPES[scanType]} SCAN`}</BoldText>
                                        <div className="details-info-time">
                                            <Icon name={ICON_NAMES.CLOCK} />
                                            <span><BoldText>Started</BoldText>{` on ${formatDate(startTime)}`}</span>
                                            {!!endTime && <span>{` - `}<BoldText>Completed</BoldText>{` on ${formatDate(endTime)}`}</span>}
                                        </div>
                                    </div>
                                    <div className="namespaces-scan-details">
                                        <span>{`with target namespaces: `}</span>
                                        {isEmpty(scannedNamespaces) ? <BoldText style={{marginLeft: "10px"}}>All</BoldText> : 
                                            <div className="namespaces-list">
                                                {scannedNamespaces.map((namespace, index) => <div key={index} className="namespace-item">{namespace}</div>)}
                                            </div>
                                        }
                                    </div>
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