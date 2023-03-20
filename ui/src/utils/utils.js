import moment from 'moment';
import { get } from 'lodash';
import { FINDINGS_MAPPING, VULNERABILITIES_ICON_NAME } from 'utils/systemConsts';
import IconWithTooltip from 'components/IconWithTooltip';
import VulnerabilitiesDisplay from 'components/VulnerabilitiesDisplay';

export const formatDateBy = (date, format) => !!date ? moment(date).format(format): "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const toCapitalized = string => string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();

export const BoldText = ({children, style={}}) => <span style={{fontWeight: "bold", ...style}}>{children}</span>;

export const getScanName = ({name, startTime}) => `${name} ${formatDate(startTime)}`;

export const getFindingsColumnsConfigList = ({tableTitle, summaryKey="summary"}) => Object.keys(FINDINGS_MAPPING).map(findingKey => {
    const {dataKey, title, icon} = FINDINGS_MAPPING[findingKey];

    return {
        Header: <IconWithTooltip tooltipId={`table-header-${tableTitle}-${dataKey}`} tooltipText={title} name={icon} />,
        id: dataKey,
        accessor: original => {
            const summary  = get(original, summaryKey) || {};
    
            return summary[dataKey] || 0;
        },
        width: 50,
        disableSort: true
    }
});

export const getVulnerabilitiesColumnConfigItem = ({tableTitle, idKey="id", summaryKey="summary"}) => ({
    Header: (
        <IconWithTooltip
            tooltipId={`table-header-${tableTitle}-vulnerabilities`}
            tooltipText="Vulnerabilities"
            name={VULNERABILITIES_ICON_NAME}
        />
    ),
    id: "vulnerabilities",
    Cell: ({row}) => {
        const id  = get(row.original, idKey) || {};
        const summary  = get(row.original, summaryKey) || {};
        
        return (
            <VulnerabilitiesDisplay minimizedTooltipId={id} counters={summary?.totalVulnerabilities} isMinimized />
        )
    },
    width: 50,
    disableSort: true
});