import React from 'react';
import VulnerabilitiesSummaryDisplay from 'components/VulnerabilitiesSummaryDisplay';
import { ICON_NAMES } from 'components/Icon';
import { CIS_SEVERITY_ITEMS } from 'utils/systemConsts';
import StepDisplay, { StepDisplayTitle } from '../StepDisplay';

import './total-display-step.scss';

const TotalDisplayStep = ({vulnerabilityPerSeverity, cisDockerBenchmarkCountPerLevel}) => (
    <StepDisplay step="2"  title="Total vulnerabilities:" className="total-display-step">
        <VulnerabilitiesSummaryDisplay id="runtime-scan-vulnerabilities" vulnerabilities={vulnerabilityPerSeverity || []} />
        <div className="cis-benchmark-total-display">
            <StepDisplayTitle>CIS Benchmark</StepDisplayTitle>
            <VulnerabilitiesSummaryDisplay
                id="runtime-scan-cis"
                vulnerabilities={cisDockerBenchmarkCountPerLevel || []}
                icon={ICON_NAMES.ALERT}
                severityItems={CIS_SEVERITY_ITEMS}
                severityKey="level"
            />
        </div>
    </StepDisplay>
)

export default TotalDisplayStep;