import React, { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import VerticalItemsList from 'components/VerticalItemsList';
import { FILTERR_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS } from 'utils/systemConsts';
import { VulnerabilitiesLink, PackagesLink, ApplicationsLink } from './utils';

const RESOURCE_TYPE_ITEMS = [
    {value: "IMAGE", label: "Image"},
    {value: "DIRECTORY", label: "Directory"},
    {value: "FILE", label: "File"}
];

const ApplicationResourcesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Resource Name",
            id: "resourceName",
            accessor: "resourceName"
        },
        {
            Header: "Resource Hash",
            id: "resourceHash",
            accessor: "resourceHash"
        },
        {
            Header: "Resource Type",
            id: "resourceType",
            accessor: "resourceType"
        },
        {
            Header: "Vulnerabilities",
            id: "vulnerabilities",
            Cell: ({row}) => {
                const {id, vulnerabilities, resourceName} = row.original;
                
                return (
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationResourceID={id} resourceName={resourceName} />
                )
            },
            width: 200,
            canSort: true
        },
        {
            Header: "Application",
            id: "applications",
            Cell: ({row}) => {
                const {id, applications, resourceName} = row.original;
                
                return (
                    <ApplicationsLink applications={applications} applicationResourceID={id} resourceName={resourceName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "Packages",
            id: "packages",
            Cell: ({row}) => {
                const {id, packages, resourceName} = row.original;
                
                return (
                    <PackagesLink packages={packages} applicationResourceID={id} resourceName={resourceName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "SBOM Analyzers",
            id: "reportingSBOMAnalyzers",
            Cell: ({row}) => <VerticalItemsList items={row.original.reportingSBOMAnalyzers} />
        }
    ], []);

    const {pathname} = useLocation();
    const navigate = useNavigate();

    return (
        <TablePage
            columns={columns}
            filterType={FILTERR_TYPES.APPLICATION_RESOURCES}
            filtersMap={{
                resourceName: {value: "resourceName", label: "Resource name", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                resourceHash: {value: "resourceHash", label: "Resource hash", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                resourceType: {value: "resourceType", label: "Resource type", operators: [
                    {...OPERATORS.is, valueItems: RESOURCE_TYPE_ITEMS, creatable: false}
                ]},
                vulnerabilitySeverity: {value: "vulnerabilitySeverity", label: "Vulnerability severity", operators: [
                    {...OPERATORS.gte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                    {...OPERATORS.lte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                ]},
                applications: {value: "applications", label: "Applications", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                packages: {value: "packages", label: "Packages", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                reportingSBOMAnalyzers: {value: "reportingSBOMAnalyzers", label: "Reporting SBOM Analyzers", operators: [
                    {...OPERATORS.containElements, valueItems: [], creatable: true},
                    {...OPERATORS.doesntContainElements, valueItems: [], creatable: true}
                ]}
            }}
            url="applicationResources"
            title="Application Resources"
            defaultSortBy={[{id: "resourceName", desc: true}]}
            onLineClick={({id}) => navigate(`${pathname}/${id}`)}
        />
    )
}

export default ApplicationResourcesTable;