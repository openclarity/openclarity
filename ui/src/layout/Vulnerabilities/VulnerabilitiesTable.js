import React, { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import VerticalItemsList from 'components/VerticalItemsList';
import SeverityTag from 'components/SeverityTag';
import InfoIcon from 'components/InfoIcon';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS } from 'utils/systemConsts';
import { PackagesLink, ApplicationsLink, ApplicationResourcesLink, CvssScoreMessage } from './utils';

const HAS_VERSION_ITEMS = [
    {value: "true", label: "available"},
    {value: "false", label: "not available"},
];

const VULNERABILITY_SOURCE_ITEMS = [
    {value: "CICD", label: "CI/CD"},
    {value: "RUNTIME", label: "Runtime"}
];

const VulnerabilitiesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Vulnerability Name",
            id: "vulnerabilityName",
            accessor: "vulnerabilityName"
        },
        {
            Header: "Severity",
            id: "severity",
            Cell: ({row}) => {
                const {id, severity, cvssSeverity, cvssBaseScore} = row.original;
                
                return (
                    <div className="vulnerability-score-wrapper">
                        <SeverityTag severity={severity} />
                        <div style={{marginTop: "6px"}}>
                            <span style={{marginRight: "6px"}}>{`CVSS: ${cvssBaseScore || "N/A"}`}</span>
                            {(!!cvssSeverity && cvssSeverity !== severity) && 
                                <InfoIcon
                                    tooltipId={`cvss-message-tooltip-${id}`}
                                    tooltipText={<div style={{width: "260px"}}><CvssScoreMessage cvssScore={cvssBaseScore} cvssSeverity={cvssSeverity} /></div>}
                                />
                            }
                        </div>
                    </div>
                )
            },
            canSort: true
        },
        {
            Header: "Package Name",
            id: "packageName",
            accessor: "packageName"
        },
        {
            Header: "Package Version",
            id: "packageVersion",
            Cell: ({row}) => {
                const {packageVersion, packageName} = row.original;
                
                return (
                    <PackagesLink packageVersion={packageVersion} packageName={packageName} />
                )
            },
            canSort: true
        },
        {
            Header: "Fix Version",
            id: "fixVersion",
            accessor: "fixVersion"
        },
        {
            Header: "Applications",
            id: "applications",
            Cell: ({row}) => {
                const {vulnerabilityName, packageID, applications} = row.original;
                
                return (
                    <ApplicationsLink packageID={packageID} applications={applications} vulnerabilityName={vulnerabilityName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "Application Resources",
            id: "applicationResources",
            Cell: ({row}) => {
                const {vulnerabilityName, packageID, applicationResources} = row.original;
                
                return (
                    <ApplicationResourcesLink packageID={packageID} applicationResources={applicationResources} vulnerabilityName={vulnerabilityName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "Scanners",
            id: "reportingScanners",
            Cell: ({row}) => <VerticalItemsList items={row.original.reportingScanners} />
        },
        {
            Header: "Source",
            id: "source",
            accessor: "source"
        }
    ], []);

    const {pathname} = useLocation();
    const navigate = useNavigate();

    return (
        <TablePage
            columns={columns}
            filterType={FILTER_TYPES.VULNERABILITIES}
            filtersMap={{
                vulnerabilityName: {value: "vulnerabilityName", label: "Vulnerability name", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                packageName: {value: "packageName", label: "Package name", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                packageVersion: {value: "packageVersion", label: "Package version", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                vulnerabilitySeverity: {value: "vulnerabilitySeverity", label: "Vulnerability severity", operators: [
                    {...OPERATORS.is, valueItems: Object.values(SEVERITY_ITEMS), creatable: false},
                    {...OPERATORS.isNot, valueItems: Object.values(SEVERITY_ITEMS), creatable: false},
                    {...OPERATORS.gte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                    {...OPERATORS.lte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                ]},
                applicationResources: {value: "applicationResources", label: "Application resources", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                applications: {value: "applications", label: "Applications", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                hasFixVersion: {value: "hasFixVersion", label: "Fix version", valuesMapItems: HAS_VERSION_ITEMS, operators: [
                    {...OPERATORS.is, valueItems: HAS_VERSION_ITEMS, creatable: false, isSingleSelect: true},
                ]},
                reportingScanners: {value: "reportingScanners", label: "Reporting scanners", operators: [
                    {...OPERATORS.containElements, valueItems: [], creatable: true},
                    {...OPERATORS.doesntContainElements, valueItems: [], creatable: true}
                ]},
                vulnerabilitySource: {value: "vulnerabilitySource", label: "Vulnerability source", operators: [
                    {...OPERATORS.is, valueItems: VULNERABILITY_SOURCE_ITEMS, creatable: false}
                ]}
            }}
            url="vulnerabilities"
            title="Vulnerabilities"
            defaultSortBy={[{id: "severity", desc: true}]}
            onLineClick={({vulnerabilityID, packageID}) => navigate(`${pathname}/${vulnerabilityID}/${packageID}`)}
        />
    )
}

export default VulnerabilitiesTable;