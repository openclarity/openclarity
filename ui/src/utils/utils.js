import moment from 'moment';
import cronstrue from 'cronstrue';
import { isNull } from 'lodash';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import IconWithTooltip from 'components/IconWithTooltip';
import VulnerabilitiesDisplay from 'components/VulnerabilitiesDisplay';

export const formatDateBy = (date, format) => !!date ? moment(date).format(format): "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const toCapitalized = string => string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();

export const BoldText = ({children, style={}}) => <span style={{fontWeight: "bold", ...style}}>{children}</span>;

export const cronExpressionToHuman = value => cronstrue.toString(value, {use24HourTimeFormat: true});

export const getScanName = ({name, startTime}) => `${name} ${formatDate(startTime)}`;

export const getFindingsColumnsConfigList = (tableTitle) => Object.keys(FINDINGS_MAPPING).map(findingKey => {
    const {totalKey, title, icon} = FINDINGS_MAPPING[findingKey];

    return {
        Header: <IconWithTooltip tooltipId={`table-header-${tableTitle}-${totalKey}`} tooltipText={title} name={icon} />,
        id: totalKey,
        accessor: original => {
            const {summary}  = original;
    
            return !isNull(summary) ? 0 : (summary[totalKey] || 0);
        },
        width: 50,
        disableSort: true
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
        Cell: ({row}) => {
            const {id, summary} = row.original;
            
            return (
                <VulnerabilitiesDisplay minimizedTooltipId={id} counters={summary?.totalVulnerabilities} isMinimized />
            )
        },
        width: 50,
        disableSort: true
    }
};