import React, { useCallback, useMemo, useState } from 'react';
import { isNull } from 'lodash';
import ButtonWithIcon from 'components/ButtonWithIcon';
import { ICON_NAMES } from 'components/Icon';
import EmptyDisplay from 'components/EmptyDisplay';
import ExpandableList from 'components/ExpandableList';
import TablePage from 'components/TablePage';
import { BoldText, toCapitalized, formatDate } from 'utils/utils';
import { APIS } from 'utils/systemConsts';
import { formatTagsToStringInstances } from 'layout/Scans/utils';
import { ExpandableScopeDisplay } from 'layout/Scans/scopeDisplayUtils';
import { useModalDisplayDispatch, MODAL_DISPLAY_ACTIONS } from 'layout/Scans/ScanConfigWizardModal/ModalDisplayProvider';
import { FILTER_TYPES } from 'context/FiltersProvider';
import ConfigurationActionsDisplay from '../ConfigurationActionsDisplay';

import './configurations-table.scss';

const TABLE_TITLE = "scan configurations";

const ConfigurationsTable = () => {
    const modalDisplayDispatch = useModalDisplayDispatch();
    const setScanConfigFormData = (data) => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA, payload: data});

    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "name",
            accessor: "name",
            disableSort: true
        },
        {
            Header: "Scope",
            id: "scope",
            Cell: ({row}) => {
                const {allRegions, regions} = row.original.scope;

                return (
                    <ExpandableScopeDisplay all={allRegions} regions={regions} />
                )
            },
            alignToTop: true,
            disableSort: true
        },
        {
            Header: "Excluded instances",
            id: "instanceTagExclusion",
            Cell: ({row}) => {
                const {instanceTagExclusion} = row.original.scope;
                
                return (
                    <ExpandableList items={formatTagsToStringInstances(instanceTagExclusion)} withTagWrap />
                )
            },
            alignToTop: true,
            disableSort: true
        },
        {
            Header: "Included instances",
            id: "instanceTagSelector",
            Cell: ({row}) => {
                const {instanceTagSelector} = row.original.scope;
                
                return (
                    <ExpandableList items={formatTagsToStringInstances(instanceTagSelector)} withTagWrap />
                )
            },
            alignToTop: true,
            disableSort: true
        },
        {
            Header: "Time config",
            id: "timeConfig",
            Cell: ({row}) => {
                const {operationTime} = row.original.scheduled;
                const isScheduled = (Date.now() - (new Date(operationTime)).valueOf() <= 0);
                
                return (
                    <div>
                        {!!isScheduled && <BoldText>Scheduled</BoldText>}
                        <div>{formatDate(operationTime)}</div>
                    </div>
                )
            },
            disableSort: true
        },
        {
            Header: "Scan types",
            id: "scanTypes",
            Cell: ({row}) => {
                const {scanFamiliesConfig} = row.original;

                return (
                    <div>
                        {
                            Object.keys(scanFamiliesConfig).map(type => {
                                const {enabled} = scanFamiliesConfig[type];

                                return enabled ? toCapitalized(type) : null;
                            }).filter(type => !isNull(type)).join(" - ")
                        }
                    </div>
                )
            },
            disableSort: true
        }
    ], []);

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = useCallback(() => setRefreshTimestamp(Date()), []);

    return (
        <div className="scan-configs-table-page-wrapper">
            <ButtonWithIcon iconName={ICON_NAMES.PLUS} onClick={() => setScanConfigFormData({})}>
                New scan configuration
            </ButtonWithIcon>
            <TablePage
                columns={columns}
                url={APIS.SCAN_CONFIGS}
                tableTitle={TABLE_TITLE}
                filterType={FILTER_TYPES.SCAN_CONFIGURATIONS}
                refreshTimestamp={refreshTimestamp}
                actionsColumnWidth={100}
                actionsComponent={({original}) => (
                    <ConfigurationActionsDisplay
                        data={original}
                        setScanConfigFormData={setScanConfigFormData}
                        onDelete={() => doRefreshTimestamp()}
                    />
                )}
                customEmptyResultsDisplay={() => (
                    <EmptyDisplay
                        message={(
                            <>
                                <div>No scan configurations detected.</div>
                                <div>Create your first scan configuration to see your VM's issues.</div>
                            </>
                        )}
                        title="New scan configuration"
                        onClick={() => setScanConfigFormData({})}
                    />
                )}
                absoluteSystemBanner
            />
        </div>
    )
}

export default ConfigurationsTable;