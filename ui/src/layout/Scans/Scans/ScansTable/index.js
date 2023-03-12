import React, { useMemo } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import ContentContainer from 'components/ContentContainer';
import EmptyDisplay from 'components/EmptyDisplay';
import Table, { utils } from 'components/Table';
import { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';
import ProgressBar from 'components/ProgressBar';
import { ExpandableScopeDisplay } from 'layout/Scans/scopeDisplayUtils';
import { SCAN_CONFIGS_PATH } from 'layout/Scans/Configurations';
import { useModalDisplayDispatch, MODAL_DISPLAY_ACTIONS } from 'layout/Scans/ScanConfigWizardModal/ModalDisplayProvider';
import { APIS } from 'utils/systemConsts';
import { formatDate } from 'utils/utils';
import VulnerabilitiesDisplay from '../VulnerabilitiesDisplay';
// import ScanActionsDisplay from '../ScanActionsDisplay';
import { FINDINGS_MAPPING, findProgressStatusFromScanState } from '../utils';

const TABLE_TITLE = "scans";
const COUNTER_CELL_WIDTH = 50;
const TIME_CELL_WIDTH = 110;

const TimeDisplay = ({time}) => (
    !!time ? formatDate(time) : <utils.EmptyValue />
);

const getFindingTypeColumn = ({value, title, icon}) => ({
    Header: <IconWithTooltip tooltipId={`table-header-${TABLE_TITLE}-${value}`} tooltipText={title} name={icon} />,
    id: value,
    accessor: `summary.${value}`,
    width: COUNTER_CELL_WIDTH,
    disableSort: true
})

const ScansTable = () => {
    const modalDisplayDispatch = useModalDisplayDispatch();

    const navigate = useNavigate();
    const {pathname} = useLocation();

    const columns = useMemo(() => [
        {
            Header: "Config Name",
            id: "name",
            accessor: "scanConfigSnapshot.name",
            disableSort: true
        },
        {
            Header: "Started",
            id: "startTime",
            Cell: ({row}) => <TimeDisplay time={row.original.startTime} />,
            width: TIME_CELL_WIDTH,
            disableSort: true
        },
        {
            Header: "Ended",
            id: "endTime",
            Cell: ({row}) => <TimeDisplay time={row.original.endTime} />,
            width: TIME_CELL_WIDTH,
            disableSort: true
        },
        {
            Header: "Scope",
            id: "scope",
            Cell: ({row}) => {
                const {all, regions} = row.original.scanConfigSnapshot?.scope;

                return <ExpandableScopeDisplay all={all} regions={regions || []} />
            },
            width: 260,
            disableSort: true
        },
        {
            Header: "Status",
            id: "status",
            Cell: ({row}) => {
                const {state, stateReason, summary} = row.original;
                const {jobsCompleted, jobsLeftToRun} = summary;

                const {status} = findProgressStatusFromScanState({state, stateReason});

                return (
                    <ProgressBar status={status} itemsCompleted={jobsCompleted} itemsLeft={jobsLeftToRun} width="80px" />
                )
            },
            width: 150,
            disableSort: true
        },
        {
            Header: <IconWithTooltip tooltipId={`table-header-${TABLE_TITLE}-vulnerabilities`} tooltipText="Vulnerabilities" name={ICON_NAMES.SHIELD} />,
            id: "vulnerabilities",
            Cell: ({row}) => {
                const {id, summary} = row.original;
                
                return (
                    <VulnerabilitiesDisplay id={id} counters={summary?.totalVulnerabilities} isMinimized />
                )
            },
            width: COUNTER_CELL_WIDTH,
            disableSort: true
        },
        getFindingTypeColumn(FINDINGS_MAPPING.EXPLOITS),
        getFindingTypeColumn(FINDINGS_MAPPING.MISCONFIGURATIONS),
        getFindingTypeColumn(FINDINGS_MAPPING.SECRETS),
        getFindingTypeColumn(FINDINGS_MAPPING.MALWARE),
        getFindingTypeColumn(FINDINGS_MAPPING.ROOTKITS),
        getFindingTypeColumn(FINDINGS_MAPPING.PACKAGES),
        {
            Header: "Scanned assets",
            id: "assets",
            accessor: original => {
                const {jobsCompleted, jobsLeftToRun} = original.summary;
                
                return `${jobsCompleted}/${jobsCompleted + jobsLeftToRun}`;
            },
            disableSort: true
        }
    ], []);

    return (
        <div className="scans-table-page-wrapper">
            <ContentContainer>
                <Table
                    columns={columns}
                    paginationItemsName={TABLE_TITLE.toLowerCase()}
                    url={APIS.SCANS}
                    noResultsTitle={TABLE_TITLE}
                    onLineClick={({id}) => navigate(`${pathname}/${id}`)}
                    // actionsComponent={({original}) => (
                    //     <ScanActionsDisplay data={original} />
                    // )}
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
                            onSubClick={() => navigate(SCAN_CONFIGS_PATH)}
                        />
                    )}
                />
            </ContentContainer>
        </div>
    )
}

export default ScansTable;