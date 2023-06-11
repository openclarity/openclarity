import React, { useMemo, useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { isNull } from 'lodash';
import { useDelete, usePrevious } from 'hooks';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import VerticalItemsList from 'components/VerticalItemsList';
import Icon, { ICON_NAMES } from 'components/Icon';
import { TooltipWrapper } from 'components/Tooltip';
import Modal from 'components/Modal';
import { CisBenchmarkLevelsDisplay } from 'components/VulnerabilitiesSummaryDisplay';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS, CIS_SEVERITY_ITEMS } from 'utils/systemConsts';
import { RESOURCE_TYPES, VulnerabilitiesLink, PackagesLink, ApplicationsLink } from './utils';

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
            Header: "CIS Docker Benchmark",
            id: "cisDockerBenchmarkResults",
            Cell: ({row}) => {
                const {id, cisDockerBenchmarkResults} = row.original;
                
                return (
                    <CisBenchmarkLevelsDisplay id={id} levels={cisDockerBenchmarkResults} withTotal />
                )
            },
            width: 150,
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
            width: 55,
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

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const [deleteConfirmationId, setDeleteConfirmationId] = useState(null);
    const closeDeleteConfirmationModal = () => setDeleteConfirmationId(null);

    const [{deleting, error: deleteError}, deleteApplication] = useDelete("applicationResources");
    const prevDeleting = usePrevious(deleting);

    useEffect(() => {
        if (prevDeleting && !deleting && !deleteError) {
            doRefreshTimestamp();
        }
    }, [prevDeleting, deleting, deleteError]);

    return (
        <>
            <TablePage
                columns={columns}
                filterType={FILTER_TYPES.APPLICATION_RESOURCES}
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
                        {...OPERATORS.is, valueItems: Object.values(RESOURCE_TYPES), creatable: false}
                    ]},
                    vulnerabilitySeverity: {value: "vulnerabilitySeverity", label: "Vulnerability severity", operators: [
                        {...OPERATORS.gte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                        {...OPERATORS.lte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                    ]},
                    cisDockerBenchmarkLevel: {value: "cisDockerBenchmarkLevel", label: "CIS Docker Benchmark level", operators: [
                        {...OPERATORS.gte, valueItems: Object.values(CIS_SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                        {...OPERATORS.lte, valueItems: Object.values(CIS_SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
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
                defaultSortBy={[{id: "vulnerabilities", desc: true}]}
                onLineClick={({id}) => navigate(`${pathname}/${id}`)}
                actionsComponent={({original}) => {
                    const {id} = original;

                    return (
                        <TooltipWrapper tooltipId={`${id}-delete`} tooltipText="Delete Application Resource" >
                            <Icon
                                name={ICON_NAMES.DELETE}
                                onClick={event => {
                                    event.stopPropagation();
                                    event.preventDefault();

                                    setDeleteConfirmationId(id);
                                }}
                            />
                        </TooltipWrapper>
                    );
                }}
                refreshTimestamp={refreshTimestamp}
            />
            {!isNull(deleteConfirmationId) &&
                <Modal
                    title="Are you sure?"
                    height={225}
                    width={512}
                    onClose={closeDeleteConfirmationModal}
                    onDone={() => {
                        deleteApplication(deleteConfirmationId);
                        closeDeleteConfirmationModal();
                    }} 
                    doneTitle="Delete"
                >
                    This operation cannot be reverted
                </Modal>
            }
        </>
    )
}

export default ApplicationResourcesTable;