import React, { useCallback, useMemo, useState } from 'react';
import { isNull } from 'lodash';
import ButtonWithIcon from 'components/ButtonWithIcon';
import { ICON_NAMES } from 'components/Icon';
import EmptyDisplay from 'components/EmptyDisplay';
import TablePage from 'components/TablePage';
import { OPERATORS } from 'components/Filter';
import { BoldText, toCapitalized, formatDate } from 'utils/utils';
import { APIS } from 'utils/systemConsts';
import { getScanTimeTypeTag } from 'layout/Scans/utils';
import { useModalDisplayDispatch, MODAL_DISPLAY_ACTIONS } from 'layout/Scans/ScanConfigWizardModal/ModalDisplayProvider';
import { FILTER_TYPES } from 'context/FiltersProvider';
import ConfigurationActionsDisplay from '../ConfigurationActionsDisplay';

import './configurations-table.scss';

const TABLE_TITLE = "scan configurations";

const SCAN_TYPES_FILTER_ITEMS = [
    "vulnerabilities",
    "exploits",
    "malware",
    "misconfigurations",
    "rootkits",
    "secrets",
    "infoFinder",
    "sbom"
].map(type => ({value: `scanTemplate.assetScanTemplate.scanFamiliesConfig.${type}.enabled`, label: toCapitalized(type)}));

const formatScanTypesToOdata = (valuesList, operator) => (
    valuesList.map(value => `(${value} eq ${operator === OPERATORS.contains.value ? "true" : "false"})`).join(` or `)
)

const ConfigurationsTable = () => {
    const modalDisplayDispatch = useModalDisplayDispatch();
    const setScanConfigFormData = (data) => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA, payload: data});

    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "name",
            sortIds: ["name"],
            accessor: "name"
        },
        {
            Header: "Scope",
            id: "scanTemplate.scope",
            sortIds: ["scanTemplate.scope"],
            accessor: "scanTemplate.scope"
        },
        {
            Header: "Scan time",
            id: "timeConfig",
            sortIds: ["scheduled.operationTime"],
            Cell: ({row}) => {
                const {operationTime, cronLine} = row.original.scheduled;
                const scanType = getScanTimeTypeTag({operationTime, cronLine});
                
                return (
                    <div>
                        {!!scanType && <BoldText>{scanType}</BoldText>}
                        <div>{formatDate(operationTime)}</div>
                    </div>
                )
            }
        },
        {
            Header: "Scan types",
            id: "scanTypes",
            sortIds: [
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.exploits.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.malware.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.misconfigurations.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.rootkits.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.sbom.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.secrets.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.vulnerabilities.enabled",
                "scanTemplate.assetScanTemplate.scanFamiliesConfig.infoFinder.enabled"
            ],
            Cell: ({row}) => {
                const {scanFamiliesConfig} = row.original.scanTemplate.assetScanTemplate;

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
            }
        }
    ], []);
    
    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = useCallback(() => setRefreshTimestamp(Date()), []);

    return (
        <div className="scan-configs-table-page-wrapper">
            <TablePage
                columns={columns}
                url={APIS.SCAN_CONFIGS}
                select="id,name,scanTemplate/scope,scheduled,scanTemplate/assetScanTemplate/scanFamiliesConfig"
                tableTitle={TABLE_TITLE}
                filterType={FILTER_TYPES.SCAN_CONFIGURATIONS}
                filtersConfig={[
                    {value: "name", label: "Name", operators: [
                        {...OPERATORS.eq, valueItems: [], creatable: true},
                        {...OPERATORS.ne, valueItems: [], creatable: true},
                        {...OPERATORS.startswith},
                        {...OPERATORS.endswith},
                        {...OPERATORS.contains, valueItems: [], creatable: true}
                    ]},
                    {value: "scanTemplate.scope", label: "Scope", operators: [
                        {...OPERATORS.eq, valueItems: [], creatable: true},
                        {...OPERATORS.ne, valueItems: [], creatable: true},
                        {...OPERATORS.startswith},
                        {...OPERATORS.endswith},
                        {...OPERATORS.contains, valueItems: [], creatable: true}
                    ]},
                    {value: "scheduled.operationTime", label: "Scan time", isDate: true, operators: [
                        {...OPERATORS.ge},
                        {...OPERATORS.le},
                    ]},
                    {value: "scanTypes", label: "Scan types", customOdataFormat: formatScanTypesToOdata, operators: [
                        {...OPERATORS.contains, valueItems: SCAN_TYPES_FILTER_ITEMS},
                        {...OPERATORS.notcontains, valueItems: SCAN_TYPES_FILTER_ITEMS}
                    ]}
                ]}
                customHeaderDisplay={() => (
                    <ButtonWithIcon className="new-config-button" iconName={ICON_NAMES.PLUS} onClick={() => setScanConfigFormData({})}>
                        New scan configuration
                    </ButtonWithIcon>
                )}
                refreshTimestamp={refreshTimestamp}
                actionsColumnWidth={130}
                actionsComponent={({original}) => (
                    <ConfigurationActionsDisplay
                        data={original}
                        setScanConfigFormData={setScanConfigFormData}
                        onDelete={doRefreshTimestamp}
                        onUpdate={doRefreshTimestamp}
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
