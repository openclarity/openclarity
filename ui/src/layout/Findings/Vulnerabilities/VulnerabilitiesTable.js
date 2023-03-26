import React, { useMemo } from 'react';
import ExpandableList from 'components/ExpandableList';
import { getAssetAndScanColumnsConfigList } from 'layout/Findings/utils';
import FindingsTablePage from '../FindingsTablePage';
import SeverityWithCvssDisplay from './SeverityWithCvssDisplay';
import { getHigestVersionCvssData } from './utils';

const VulnerabilitiesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Vulnerability name",
            id: "name",
            accessor: "findingInfo.vulnerabilityName",
            disableSort: true
        },
        {
            Header: "Severity",
            id: "severity",
            Cell: ({row}) => {
                const {id, findingInfo} = row.original;
                const {severity, cvss} = findingInfo || {};
                const cvssScoreData = getHigestVersionCvssData(cvss);
                
                return (
                    <SeverityWithCvssDisplay
                        severity={severity}
                        cvssScore={cvssScoreData?.score}
                        cvssSeverity={cvssScoreData?.severity?.toLocaleUpperCase()}
                        compareTooltipId={`severity-compare-tooltip-${id}`}
                    />
                )
            },
            disableSort: true
        },
        {
            Header: "Package name",
            id: "packageName",
            accessor: "findingInfo.package.name",
            disableSort: true
        },
        {
            Header: "Package version",
            id: "packageVersion",
            accessor: "findingInfo.package.version",
            disableSort: true
        },
        {
            Header: "Fix versions",
            id: "fixVersions",
            Cell: ({row}) => {
                const {versions} = row.original.findingInfo?.fix || {};

                return (
                    <ExpandableList items={versions || []} />
                )
            },
            disableSort: true
        },
        ...getAssetAndScanColumnsConfigList()
    ], []);

    return (
        <FindingsTablePage
            columns={columns}
            tableTitle="vulnerabilities"
            findingsObjectType="Vulnerability"
        />
    )
}

export default VulnerabilitiesTable;