import { orderBy } from 'lodash';
import CVSS from '@turingpointde/cvss.js';

export const getHigestVersionCvssData = (cvssData) => {
    const sortedCvss = orderBy(cvssData || [], ["version"], ["desc"]);

    const {vector, metrics, version} = sortedCvss[0];

    const serverData = {
        vector,
        score: metrics.baseScore,
        exploitabilityScore: metrics.exploitabilityScore,
        impactScore: metrics.impactScore
    }

    if (version === "2.0") {
        return serverData
    }

    const cvssVector = CVSS(vector);

    return {
        ...serverData,
        temporalScore: cvssVector.getTemporalScore(),
        environmentalScore: cvssVector.getEnvironmentalScore(),
        severity: cvssVector.getRating(),
        metrics: cvssVector.getDetailedVectorObject().metrics,
    }
}