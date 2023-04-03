import moment from 'moment';
import cronstrue from 'cronstrue';
import CVSS from '@turingpointde/cvss.js';
import { isEmpty, orderBy } from 'lodash';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import IconWithTooltip from 'components/IconWithTooltip';
import VulnerabilitiesDisplay from 'components/VulnerabilitiesDisplay';

export const formatDateBy = (date, format) => !!date ? moment(date).format(format): "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const calculateDuration = (startTime, endTime) => {
    const startMoment = moment(startTime);
    const endMoment = moment(endTime);
    
    if (!startTime || !endTime) {
        return null;
    }

    const range = ["days", "hours", "minutes", "seconds"].map(item => ({diff: endMoment.diff(startMoment, item), label: item}))
        .find(({diff}) => diff > 1);

    return !!range ? `${range.diff} ${range.label}` : null;
}

export const toCapitalized = string => string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();

export const BoldText = ({children, style={}}) => <span style={{fontWeight: "bold", ...style}}>{children}</span>;

export const cronExpressionToHuman = value => cronstrue.toString(value, {use24HourTimeFormat: true});

export const formatNumber = value => (
    new Intl.NumberFormat("en-US").format(parseInt(value || 0, 10))
)

export const getScanName = ({name, startTime}) => `${name} ${formatDate(startTime)}`;

export const getHigestVersionCvssData = (cvssData) => {
    if (isEmpty(cvssData)) {
        return {};
    }
    
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

export const getFindingsColumnsConfigList = (tableTitle) => Object.keys(FINDINGS_MAPPING).map(findingKey => {
    const {totalKey, title, icon} = FINDINGS_MAPPING[findingKey];

    return {
        Header: <IconWithTooltip tooltipId={`table-header-${tableTitle}-${totalKey}`} tooltipText={title} name={icon} />,
        id: totalKey,
        sortIds: [`summary.${totalKey}`],
        accessor: original => {
            const {summary}  = original;
            
            return isEmpty(summary) ? 0 : (formatNumber(summary[totalKey] || 0));
        },
        width: 50
    }
});

export const getVulnerabilitiesColumnConfigItem = (tableTitle) => {
    const {title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

    return {
        Header: (
            <IconWithTooltip
                tooltipId={`table-header-${tableTitle}-vulnerabilities`}
                tooltipText={vulnerabilitiesTitle}
                name={vulnerabilitiesIcon}
            />
        ),
        id: "vulnerabilities",
        sortIds: [
            "summary.totalVulnerabilities.totalCriticalVulnerabilities",
            "summary.totalVulnerabilities.totalHighVulnerabilities",
            "summary.totalVulnerabilities.totalMediumVulnerabilities",
            "summary.totalVulnerabilities.totalLowVulnerabilities",
            "summary.totalVulnerabilities.totalNegligibleVulnerabilities"
        ],
        Cell: ({row}) => {
            const {id, summary} = row.original;
            
            return (
                <VulnerabilitiesDisplay minimizedTooltipId={id} counters={summary?.totalVulnerabilities} isMinimized />
            )
        },
        width: 50
    }
};