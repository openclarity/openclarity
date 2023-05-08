import React, { useMemo, useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { isNull } from 'lodash';
import { useDelete, usePrevious } from 'hooks';
import TablePage, { TableHeaderPortal } from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import VerticalItemsList from 'components/VerticalItemsList';
import Icon, { ICON_NAMES } from 'components/Icon';
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import Modal from 'components/Modal';
import { TooltipWrapper } from 'components/Tooltip';
import { LabelsDisplay } from 'components/LabelTag';
import { CisBenchmarkLevelsDisplay } from 'components/VulnerabilitiesSummaryDisplay';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { SEVERITY_ITEMS, CIS_SEVERITY_ITEMS } from 'utils/systemConsts';
import { APPLICATION_TYPE_ITEMS, VulnerabilitiesLink, PackagesLink, ApplicationResourcesLink } from './utils';
import ApplicationForm, { APP_FIELD_NAMES } from './ApplicationForm';

import './applications.scss';

const LabelsCellDisplay = (props) => {
    const labels = props.labels || [];
    const MINIMAL_LEN = 3;
    const minimalLabels = labels.slice(0, MINIMAL_LEN);

    const [labelsToDisplay, setLabelsToDisplay] = useState(labels.length > MINIMAL_LEN ? minimalLabels : labels);
    const isOpen = labelsToDisplay.length === labels.length;

    return (
        <div>
            <div className="labels-cell-display-wrapper">
                <LabelsDisplay labels={labelsToDisplay} />
                {minimalLabels.length !== labels.length &&
                    <Arrow
                        name={isOpen ? ARROW_NAMES.TOP : ARROW_NAMES.BOTTOM}
                        onClick={event => {
                            event.stopPropagation();
                            event.preventDefault();
                            
                            setLabelsToDisplay(isOpen ? minimalLabels : labels);
                        }}
                        small
                    />
                }
            </div>
        </div>
    )
}

const ApplicationsTable = () => {
    const columns = useMemo(() => [
        {
            Header: "Application Name",
            id: "applicationName",
            accessor: "applicationName"
        },
        {
            Header: "ID",
            id: "id",
            accessor: "id",
            disableSortBy: true
        },
        {
            Header: "Application Type",
            id: "applicationType",
            accessor: "applicationType"
        },
        {
            Header: "Labels",
            id: "labels",
            Cell: ({row}) => <LabelsCellDisplay labels={row.original.labels} />
        },
        {
            Header: "Environments",
            id: "environments",
            Cell: ({row}) => <VerticalItemsList items={row.original.environments} />
        },
        {
            Header: "Vulnerabilities",
            id: "vulnerabilities",
            Cell: ({row}) => {
                const {id, vulnerabilities, applicationName} = row.original;
                
                return (
                    <VulnerabilitiesLink id={id} vulnerabilities={vulnerabilities} applicationID={id} applicationName={applicationName} />
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
            width: 180,
            canSort: true
        },
        {
            Header: "Application Resources",
            id: "applicationResources",
            Cell: ({row}) => {
                const {id, applicationResources, applicationName} = row.original;
                
                return (
                    <ApplicationResourcesLink applicationResources={applicationResources} applicationID={id} applicationName={applicationName} />
                )
            },
            width: 50,
            canSort: true
        },
        {
            Header: "Packages",
            id: "packages",
            Cell: ({row}) => {
                const {id, packages, applicationName} = row.original;

                return (
                    <PackagesLink packages={packages} applicationID={id} applicationName={applicationName} />
                )
            },
            width: 50,
            canSort: true
        },
    ], []);

    const {pathname} = useLocation();
    const navigate = useNavigate();

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const [applicationData, setApplicationData] = useState(null);
    const closeApplicationForm = () => setApplicationData(null);

    const [deleteConfirmationId, setDeleteConfirmationId] = useState(null);
    const closeDeleteConfirmationModal = () => setDeleteConfirmationId(null);

    const [{deleting, error: deleteError}, deleteApplication] = useDelete("applications");
    const prevDeleting = usePrevious(deleting);

    useEffect(() => {
        if (prevDeleting && !deleting && !deleteError) {
            doRefreshTimestamp();
        }
    }, [prevDeleting, deleting, deleteError]);

    return (
        <div className="applications-table-page">
            <TableHeaderPortal>
                <div className="new-application-button" onClick={() => setApplicationData({})}>
                    <Icon name={ICON_NAMES.PLUS} />
                    <div>New Application</div>
                </div>
            </TableHeaderPortal>
            <TablePage
                columns={columns}
                filterType={FILTER_TYPES.APPLICATIONS}
                filtersMap={{
                    applicationName: {value: "applicationName", label: "Application name", operators: [
                        {...OPERATORS.is, valueItems: [], creatable: true},
                        {...OPERATORS.isNot, valueItems: [], creatable: true},
                        {...OPERATORS.start},
                        {...OPERATORS.end},
                        {...OPERATORS.contains, valueItems: [], creatable: true}
                    ]},
                    applicationType: {value: "applicationType", label: "Application type", operators: [
                        {...OPERATORS.is, valueItems: APPLICATION_TYPE_ITEMS, creatable: false},
                    ]},
                    applicationLabels: {value: "applicationLabels", label: "Application labels", operators: [
                        {...OPERATORS.containElements, valueItems: [], creatable: true},
                        {...OPERATORS.doesntContainElements, valueItems: [], creatable: true}
                    ]},
                    applicationEnvs: {value: "applicationEnvs", label: "Application environments", operators: [
                        {...OPERATORS.containElements, valueItems: [], creatable: true},
                        {...OPERATORS.doesntContainElements, valueItems: [], creatable: true}
                    ]},
                    vulnerabilitySeverity: {value: "vulnerabilitySeverity", label: "Vulnerability severity", operators: [
                        {...OPERATORS.gte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                        {...OPERATORS.lte, valueItems: Object.values(SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                    ]},
                    cisDockerBenchmarkLevel: {value: "cisDockerBenchmarkLevel", label: "CIS Docker Benchmark level", operators: [
                        {...OPERATORS.gte, valueItems: Object.values(CIS_SEVERITY_ITEMS), creatable: false, isSingleSelect: true},
                        {...OPERATORS.lte, valueItems: Object.values(CIS_SEVERITY_ITEMS), creatable: false, isSingleSelect: true}
                    ]},
                    applicationResources: {value: "applicationResources", label: "Application resources", operators: [
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
                }}
                url="applications"
                title="Applications"
                defaultSortBy={[{id: "vulnerabilities", desc: true}]}
                onLineClick={({id}) => navigate(`${pathname}/${id}`)}
                actionsComponent={({original}) => {
                    const {id} = original;
                    const deleteTooltipId = `${id}-delete`;
                    const editTooltipId = `${id}-edit`;

                    return (
                        <div className="application-row-actions">
                            <TooltipWrapper tooltipId={editTooltipId} tooltipText="Edit Application" >
                                <Icon
                                    name={ICON_NAMES.EDIT}
                                    onClick={event => {
                                        event.stopPropagation();
                                        event.preventDefault();

                                        setApplicationData(original);
                                    }}
                                />
                            </TooltipWrapper>
                            <TooltipWrapper tooltipId={deleteTooltipId} tooltipText="Delete Application" >
                                <Icon
                                    name={ICON_NAMES.DELETE}
                                    onClick={event => {
                                        event.stopPropagation();
                                        event.preventDefault();

                                        setDeleteConfirmationId(id);
                                    }}
                                />
                            </TooltipWrapper>
                        </div>
                    );
                }}
                refreshTimestamp={refreshTimestamp}
            />
            {!isNull(deleteConfirmationId) &&
                <Modal
                    title="Are you sure?"
                    className="application-delete-confirmation-modal"
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
            {!isNull(applicationData) &&
                <ApplicationForm
                    initialData={{
                        id: applicationData.id,
                        [APP_FIELD_NAMES.NAME]: applicationData.applicationName || "",
                        [APP_FIELD_NAMES.TYPE]: applicationData.applicationType || "",
                        [APP_FIELD_NAMES.LABELS]: applicationData.labels || [],
                        [APP_FIELD_NAMES.ENVIRONMENTS]: applicationData.environments || []
                    }}
                    onClose={closeApplicationForm}
                    onSuccess={() => {
                        doRefreshTimestamp();
                        closeApplicationForm();
                    }}
                />
            }
        </div>
    )
}

export default ApplicationsTable;