import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import SeverityWithCvssDisplay from 'components/SeverityWithCvssDisplay';
import { getHigestVersionCvssData } from 'utils/utils';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import FindingsTablePage from '../FindingsTablePage';

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
            tableTitle="vulnerabilities"
            findingsObjectType="Vulnerability"
        />
    )
}

export default VulnerabilitiesTable;