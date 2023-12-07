import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import SeverityWithCvssDisplay, { SEVERITY_ITEMS } from 'components/SeverityWithCvssDisplay';
import { OPERATORS } from 'components/Filter';
import { getHigestVersionCvssData, toCapitalized } from 'utils/utils';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import FindingsTablePage from '../FindingsTablePage';

const FILTER_SEVERITY_ITEMS = Object.values(SEVERITY_ITEMS).filter(({valueKey}) => valueKey !== SEVERITY_ITEMS.NONE.valueKey)
    .map(({valueKey}) => ({value: valueKey, label: toCapitalized(valueKey)}));

const FILTER_FIX_AVAILABLE_ITEMS = [
    {value: "fixed", label: "available"}
]

const VulnerabilitiesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Vulnerability name",
            id: "name",
            sortIds: ["findingInfo.vulnerabilityName"],
            accessor: "findingInfo.vulnerabilityName"
        },
        {
            Header: "Severity",
            id: "severity",
            sortIds: ["findingInfo.severity"],
            Cell: ({row}) => {
                const {id, findingInfo} = row.original;
                const {severity, cvss} = findingInfo || {};
                const cvssScoreData = getHigestVersionCvssData(cvss);
                
                return (
                    <SeverityWithCvssDisplay
                        severity={severity}
                        cvssScore={cvssScoreData?.score}
                        cvssSeverity={cvssScoreData?.severity?.toUpperCase()}
                        compareTooltipId={`severity-compare-tooltip-${id}`}
                    />
                )
            }
        },
        {
            Header: "Package name",
            id: "packageName",
            sortIds: ["findingInfo.package.name"],
            accessor: "findingInfo.package.name"
        },
        {
            Header: "Package version",
            id: "packageVersion",
            sortIds: ["findingInfo.package.version"],
            accessor: "findingInfo.package.version"
        },
        {
            Header: "Fix versions",
            id: "fixVersions",
            sortIds: ["findingInfo.fix"],
            Cell: ({row}) => {
                const {versions} = row.original.findingInfo?.fix || {};

                return (
                    <ExpandableList items={versions || []} />
                )
            }
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            filterType={FILTER_TYPES.FINDINGS_VULNERABILITIES}
            filtersConfig={[
                {value: "findingInfo.vulnerabilityName", label: "Vulnerability name", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.severity", label: "Vulnerability severity", operators: [
                    {...OPERATORS.eq, valueItems: FILTER_SEVERITY_ITEMS},
                    {...OPERATORS.ne, valueItems: FILTER_SEVERITY_ITEMS}
                ]},
                {value: "findingInfo.package.name", label: "Package name", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.package.version", label: "Package version", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                {value: "findingInfo.fix.state", label: "Fix version", operators: [
                    {...OPERATORS.eq, valueItems: FILTER_FIX_AVAILABLE_ITEMS},
                    {...OPERATORS.ne, valueItems: FILTER_FIX_AVAILABLE_ITEMS}
                ]}
            ]}
            tableTitle="vulnerabilities"
            findingsObjectType="Vulnerability"
        />
    )
}

export default VulnerabilitiesTable;
