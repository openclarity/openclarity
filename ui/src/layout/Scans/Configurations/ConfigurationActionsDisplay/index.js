import React, { useEffect, useState } from 'react';
import { isNull } from 'lodash';
import { useDelete, usePrevious, useFetch, FETCH_METHODS } from 'hooks';
import { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';
import Modal from 'components/Modal';
import { BoldText } from 'utils/utils';
import { APIS } from 'utils/systemConsts';
import { useModalDisplayDispatch, MODAL_DISPLAY_ACTIONS } from 'layout/Scans/ScanConfigWizardModal/ModalDisplayProvider';

import './configuration-actions-display.scss';

const ConfigurationActionsDisplay = ({data, onDelete, onUpdate}) => {
    const modalDisplayDispatch = useModalDisplayDispatch();
    const setScanConfigFormData = (data) => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA, payload: data});
    
    const {id, scheduled} = data;
    const {operationTime, cronLine} = scheduled || {};

    const disableStartScan = (Date.now() - (new Date(operationTime)).valueOf() <= 0) && !cronLine;

    const [deleteConfigmationData, setDeleteConfigmationData] = useState(null);
    const closeDeleteConfigmation = () => setDeleteConfigmationData(null);

    const [{deleting}, deleteConfiguration] = useDelete(APIS.SCAN_CONFIGS);
    const prevDeleting = usePrevious(deleting);

    const [{loading, error}, fetchScanConfig] = useFetch(APIS.SCAN_CONFIGS, {loadOnMount: false});
	const prevLoading = usePrevious(loading);

    useEffect(() => {
        if (prevDeleting && !deleting) {
            onDelete();
        }
    }, [prevDeleting, deleting, onDelete]);

    useEffect(() => {
        if (!!onUpdate && prevLoading && !loading && !error) {
            onUpdate();
        }
    }, [prevLoading, loading, error, onUpdate]);

    return (
        <>
            <div className="configuration-actions-display">
                <IconWithTooltip
                    tooltipId={`${id}-start`}
                    tooltipText={disableStartScan ? "A scan cannot be started for non-repetitive scan configs that are scheduled for a future time. You can edit or duplicate it to scan 'Now'." : "Start scan"}
                    name={ICON_NAMES.PLAY}
                    onClick={event => {
                        event.stopPropagation();
                        event.preventDefault();
                        
                        fetchScanConfig({
                            method: FETCH_METHODS.PATCH,
                            submitData: {scheduled: {...scheduled, operationTime: (new Date()).toISOString()}, disabled: false},
                            formatUrl: url => `${url}/${id}`
                        });
                    }}
                    disabled={disableStartScan}
                />
                <IconWithTooltip
                    tooltipId={`${id}-duplicate`}
                    tooltipText="Duplicate scan configuration"
                    name={ICON_NAMES.DUPLICATE}
                    onClick={event => {
                        event.stopPropagation();
                        event.preventDefault();
                        
                        setScanConfigFormData({...data, id: null, name: ""});
                    }}
                />
                <IconWithTooltip
                    tooltipId={`${id}-edit`}
                    tooltipText="Edit scan configuration"
                    name={ICON_NAMES.EDIT}
                    onClick={event => {
                        event.stopPropagation();
                        event.preventDefault();
                        
                        setScanConfigFormData(data);
                    }}
                />
                <IconWithTooltip
                    tooltipId={`${id}-delete`}
                    tooltipText="Delete scan configuration"
                    name={ICON_NAMES.DELETE}
                    onClick={event => {
                        event.stopPropagation();
                        event.preventDefault();

                        setDeleteConfigmationData(data);
                    }}
                />
            </div>
            {!isNull(deleteConfigmationData) &&
                <Modal
                    title="Delete configmation"
                    isMediumTitle
                    className="scan-config-delete-confirmation"
                    onClose={closeDeleteConfigmation}
                    height={250}
                    doneTitle="Delete"
                    onDone={() => {
                        deleteConfiguration(deleteConfigmationData.id);
                        closeDeleteConfigmation();
                    }}
                >
                    <span>{`Once `}</span><BoldText>{deleteConfigmationData.name}</BoldText><span>{` will be deleted, the action cannot be reverted`}</span><br />
                    <span>{`Are you sure you want to delete ${deleteConfigmationData.name}?`}</span>
                </Modal>
            }
        </>
    );
}

export default ConfigurationActionsDisplay;