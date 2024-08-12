import React, { useMemo, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import TablePage from 'components/TablePage';
import { utils } from 'components/Table';
import ScanProgressBar, { SCAN_STATES } from 'components/ScanProgressBar';
import EmptyDisplay from 'components/EmptyDisplay';
import { OPERATORS } from 'components/Filter';
import { useModalDisplayDispatch, MODAL_DISPLAY_ACTIONS } from 'layout/Scans/ScanConfigWizardModal/ModalDisplayProvider';
import { APIS, ROUTES } from 'utils/systemConsts';
import { formatDate, getFindingsColumnsConfigList, getVulnerabilitiesColumnConfigItem, formatNumber, findingsColumnsFiltersConfig,
    vulnerabilitiesCountersColumnsFiltersConfig } from 'utils/utils';
import { FILTER_TYPES } from 'context/FiltersProvider';
import ScanActionsDisplay from './ScanActionsDisplay';
import { SCANS_PATHS } from '../utils';

const TABLE_TITLE = "scans";
const TIME_CELL_WIDTH = 110;

const START_TIME_SORT_IDS = ["startTime"];

const FILTER_SCAN_STATUS_ITEMS = Object.values(SCAN_STATES).map(({state, title}) => ({value: state, label: title}));

const TimeDisplay = ({time}) => (
    !!time ? formatDate(time) : <utils.EmptyValue />
);

const ScansTable = () => {
    const modalDisplayDispatch = useModalDisplayDispatch();

    const navigate = useNavigate();

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = useCallback(() => setRefreshTimestamp(Date()), []);

    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "name",
            sortIds: ["name"],
            accessor: "name"
        },
        {
            Header: "Started",
            id: "startTime",
            sortIds: START_TIME_SORT_IDS,
            Cell: ({row}) => <TimeDisplay time={row.original.startTime} />,
            width: TIME_CELL_WIDTH
        },
        {
            Header: "Ended",
            id: "endTime",
            sortIds: ["endTime"],
            Cell: ({row}) => <TimeDisplay time={row.original.endTime} />,
            width: TIME_CELL_WIDTH
        },
        {
            Header: "Scope",
            id: "scope",
            sortIds: ["scope"],
            accessor: "scope"
        },
        {
            Header: "Status",
            id: "status",
            sortIds: ["state"],
            Cell: ({row}) => {
                const {state, reason, message} = row.original.status || {};
                const {jobsCompleted, jobsLeftToRun} = row.original.summary || {};

                return (
                    <ScanProgressBar
                        state={state}
                        stateReason={reason}
                        stateMessage={message}
                        itemsCompleted={jobsCompleted}
                        itemsLeft={jobsLeftToRun}
                        barWidth="80px"
                        isMinimized
                        minimizedTooltipId={row.original.id}
                    />
                )
            },
            width: 150
        },
        getVulnerabilitiesColumnConfigItem(TABLE_TITLE),
        ...getFindingsColumnsConfigList(TABLE_TITLE),
        {
            Header: "Scanned assets",
            id: "assets",
            sortIds: ["summary.jobsCompleted"],
            accessor: original => {
                const {jobsCompleted, jobsLeftToRun} = original.summary || {};
                
                return `${formatNumber(jobsCompleted)}/${formatNumber(jobsCompleted + jobsLeftToRun)}`;
            }
        }
    ], []);
    
    return (
        <TablePage
            columns={columns}
            url={APIS.SCANS}
            tableTitle={TABLE_TITLE}
            filterType={FILTER_TYPES.SCANS}
            filtersConfig={[
                {value: "name", label: "Name", operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueitems: [], creatable: true}
                ]},
                {value: "scope", label: "Scope", operators: [
                    {...OPERATORS.eq, valueitems: [], creatable: true},
                    {...OPERATORS.ne, valueitems: [], creatable: true},
                    {...OPERATORS.startswith},
                    {...OPERATORS.endswith},
                    {...OPERATORS.contains, valueitems: [], creatable: true}
                ]},
                {value: "startTime", label: "Started", isDate: true, operators: [
                    {...OPERATORS.ge},
                    {...OPERATORS.le},
                ]},
                {value: "endTime", label: "Ended", isDate: true, operators: [
                    {...OPERATORS.ge},
                    {...OPERATORS.le},
                ]},
                {value: "state", label: "Status", operators: [
                    {...OPERATORS.eq, valueItems: FILTER_SCAN_STATUS_ITEMS},
                    {...OPERATORS.ne, valueItems: FILTER_SCAN_STATUS_ITEMS}
                ]},
                ...vulnerabilitiesCountersColumnsFiltersConfig,
                ...findingsColumnsFiltersConfig,
                {value: "summary.jobsCompleted", label: "Scanned assets", isNumber: true, operators: [
                    {...OPERATORS.eq, valueItems: [], creatable: true},
                    {...OPERATORS.ne, valueItems: [], creatable: true},
                    {...OPERATORS.ge},
                    {...OPERATORS.le},
                ]}
            ]}
            defaultSortBy={{sortIds: START_TIME_SORT_IDS, desc: true}}
            actionsComponent={({original}) => (
                <ScanActionsDisplay data={original} onUpdate={doRefreshTimestamp} />
            )}
            customEmptyResultsDisplay={() => (
                <EmptyDisplay
                    message={(
                        <>
                            <div>No scans detected.</div>
                            <div>Start your first scan to see your VM's issues.</div>
                        </>
                    )}
                    title="New scan configuration"
                    onClick={() => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA, payload: {}})}
                    subTitle="Start scan from config"
                    onSubClick={() => navigate(`${ROUTES.SCANS}/${SCANS_PATHS.CONFIGURATIONS}`)}
                />
            )}
            refreshTimestamp={refreshTimestamp}
            absoluteSystemBanner
        />
    )
}

export default ScansTable;
