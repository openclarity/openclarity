import React from 'react';
import { FETCH_METHODS } from 'hooks';
import WizardModal from 'components/WizardModal';
import { APIS } from 'utils/systemConsts';
import StepGeneralProperties from './StepGeneralProperties';
import StepScanTypes from './StepScanTypes';
import StepTimeConfiguration, { SCHEDULE_TYPES_ITEMS, CRON_QUICK_OPTIONS } from './StepTimeConfiguration';
import StepAdvancedSettings from './StepAdvancedSettings';

import './scan-config-wizard-modal.scss';

const padDateTime = time => String(time).padStart(2, "0");

const ScanConfigWizardModal = ({initialData, onClose, onSubmitSuccess}) => {
    const {id, name, scanTemplate, scheduled} = initialData || {};
    const {scope, maxParallelScanners, assetScanTemplate} = scanTemplate || {};
    const {operationTime, cronLine} = scheduled || {};

    const {scanFamiliesConfig, scannerInstanceCreationConfig} = assetScanTemplate || {}
    const {useSpotInstances} = scannerInstanceCreationConfig || {};
    
    const isEditForm = !!id;
    
    const initialValues = {
        id: id || null,
        name: name || "",
        scanFamiliesConfig: {
            sbom: {enabled: true},
            vulnerabilities: {enabled: true},
            malware: {enabled: false},
            rootkits: {enabled: false},
            secrets: {enabled: false},
            misconfigurations: {enabled: false},
            infoFinder: {enabled: false},
            exploits: {enabled: false}
        },
        scanTemplate: {
            scope: scope || "",
            maxParallelScanners: maxParallelScanners || 2,
            assetScanTemplate: {
                scanFamiliesConfig: {
                    sbom: {enabled: true},
                    vulnerabilities: {enabled: true},
                    malware: {enabled: false},
                    rootkits: {enabled: false},
                    secrets: {enabled: false},
                    misconfigurations: {enabled: false},
                    infoFinder: {enabled: false},
                    exploits: {enabled: false}
                },
                scannerInstanceCreationConfig: {
                    useSpotInstances: useSpotInstances || false
                }
            }
        },
        scheduled: {
            scheduledSelect: !!cronLine ? SCHEDULE_TYPES_ITEMS.REPETITIVE.value : SCHEDULE_TYPES_ITEMS.NOW.value,
            laterDate: "",
            laterTime: "",
            cronLine: cronLine || CRON_QUICK_OPTIONS[0].value
        },
    }

    if (!!operationTime && !cronLine) {
        const dateTime = new Date(operationTime);
        initialValues.scheduled.scheduledSelect = SCHEDULE_TYPES_ITEMS.LATER.value;
        initialValues.scheduled.laterTime = `${padDateTime(dateTime.getHours())}:${padDateTime(dateTime.getMinutes())}`;
        initialValues.scheduled.laterDate = `${dateTime.getFullYear()}-${padDateTime(dateTime.getMonth() + 1)}-${padDateTime(dateTime.getDate())}`;
    }

    Object.keys(scanFamiliesConfig || {}).forEach(type => {
        const {enabled} = scanFamiliesConfig[type];
        initialValues.scanTemplate.assetScanTemplate.scanFamiliesConfig[type].enabled = enabled;
    })

    const steps = [
        {
            id: "general",
            title: "General properties",
            component: StepGeneralProperties
        },
        {
            id: "scanTypes",
            title: "Scan types",
            component: StepScanTypes
        },
        {
            id: "time",
            title: "Time configuration",
            component: StepTimeConfiguration
        },
        {
            id: "advance",
            title: "Advanced settings",
            component: StepAdvancedSettings
        }
    ];

    return (
        <WizardModal
            title={`${isEditForm ? "Edit" : "New"} scan config`}
            onClose={onClose}
            steps={steps}
            initialValues={initialValues}
            submitUrl={APIS.SCAN_CONFIGS}
            getSubmitParams={formValues => {
                const {id, scheduled, ...submitData} = formValues;

                const {scheduledSelect, laterDate, laterTime, cronLine} = scheduled;
                const isNow = scheduledSelect === SCHEDULE_TYPES_ITEMS.NOW.value;
                
                let formattedDate = new Date();

                if (!isNow) {
                    const [hours, minutes] = laterTime.split(":");
                    formattedDate = new Date(laterDate);
                    formattedDate.setHours(hours, minutes);
                }

                submitData.scheduled = {};

                if (scheduledSelect === SCHEDULE_TYPES_ITEMS.REPETITIVE.value) {
                    submitData.scheduled.cronLine = cronLine;
                } else {
                    submitData.scheduled.operationTime = formattedDate.toISOString();
                }

                return !isEditForm ? {submitData} : {
                    method: FETCH_METHODS.PUT,
                    formatUrl: url => `${url}/${id}`,
                    submitData
                }
            }}
            onSubmitSuccess={onSubmitSuccess}
        />
    )
}

export default ScanConfigWizardModal;
