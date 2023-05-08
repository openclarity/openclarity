import React, { useMemo, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import { useFilterDispatch, resetFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS } from 'utils/systemConsts';
import { VulnerabilitiesLink, ApplicationsLink, ApplicationResourcesLink } from './utils';

const PackagesTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Package Name",
            id: "packageName",
            accessor: "packageName"
        },
        {
            Header: "Version",
            id: "version",
            accessor: "version"
        },
        {
            Header: "Language",
            id: "language",
            accessor: "language"
        },
        {
            Header: "License",
            id: "license",
            accessor: "license"
        },
        {
            Header: "Vulnerabilities",
            id: "vulnerabilities",
            Cell: ({row}) => {
                const {id, vulnerabilities, packageName, version} = row.original;
                
                return (
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} packageVersion={version} packageName={packageName} />
                )
            },
            width: 200,
            canSort: true
        },
        {
            Header: "Applications",
            id: "applications",
            Cell: ({row}) => {
                const {id, applications, packageName} = row.original;
                
                return (
                    <ApplicationsLink applications={applications} packageID={id} packageName={packageName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "Application Resources",
            id: "applicationResources",
            Cell: ({row}) => {
                const {id, applicationResources, packageName} = row.original;
                
                return (
                    <ApplicationResourcesLink applicationResources={applicationResources} packageID={id} packageName={packageName} />
                )
            },
            width: 50,
            canSort: true
        }
    ], []);

    const {pathname} = useLocation();
    const navigate = useNavigate();

    const filtersDispatch = useFilterDispatch();

    useEffect(() => {
        resetFilters(filtersDispatch, FILTER_TYPES.PACKAGE_RESOURCES);
    }, [filtersDispatch]);

    return (
        <TablePage
            columns={columns}
            filterType={FILTER_TYPES.PACKAGES}
            filtersMap={{
                packageName: {value: "packageName", label: "Package name", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                applications: {value: "applications", label: "Applications", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                applicationResources: {value: "applicationResources", label: "Application resources", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.gte},
                    {...OPERATORS.lte}
                ]},
                language: {value: "language", label: "Language", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                vulnerabilitySeverity: {value: "vulnerabilitySeverity", label: "Vulnerability severity", operators: [
                    {...OPERATORS.gte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                    {...OPERATORS.lte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                ]},
                packageVersion: {value: "packageVersion", label: "Package version", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]},
                license: {value: "license", label: "License", operators: [
                    {...OPERATORS.is, valueItems: [], creatable: true},
                    {...OPERATORS.isNot, valueItems: [], creatable: true},
                    {...OPERATORS.start},
                    {...OPERATORS.end},
                    {...OPERATORS.contains, valueItems: [], creatable: true}
                ]}
            }}
            url="packages"
            title="Packages"
            defaultSortBy={[{id: "vulnerabilities", desc: true}]}
            onLineClick={({id}) => navigate(`${pathname}/${id}`)}
        />
    )
}

export default PackagesTable;