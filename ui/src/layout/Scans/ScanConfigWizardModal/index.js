import React from 'react';
import { isEmpty } from 'lodash';
import { FETCH_METHODS } from 'hooks';
import WizardModal from 'components/WizardModal';
import { APIS } from 'utils/systemConsts';
import { formatStringInstancesToTags, formatTagsToStringInstances } from '../utils';
import StepGeneralProperties, { REGIONS_EMPTY_VALUE, VPCS_EMPTY_VALUE, SCOPE_ITEMS } from './StepGeneralProperties';
import StepScanTypes from './StepScanTypes';
import StepTimeConfiguration, { SCHEDULE_TYPES_ITEMS, CRON_QUICK_OPTIONS } from './StepTimeConfiguration';
import StepAdvancedSettings from './StepAdvancedSettings';

import './scan-config-wizard-modal.scss';

const padDateTime = time => String(time).padStart(2, "0");

const ScanConfigWizardModal = ({initialData, onClose, onSubmitSuccess}) => {
    const {id, name, scope, scanFamiliesConfig, scheduled, maxParallelScanners, scannerInstanceCreationConfig} = initialData || {};
    const {allRegions, regions, shouldScanStoppedInstances, instanceTagSelector, instanceTagExclusion} = scope || {};
    const {operationTime, cronLine} = scheduled || {};
    const {useSpotInstances} = scannerInstanceCreationConfig || {};
    
    const isEditForm = !!id;
    
    const initialValues = {
        id: id || null,
        name: name || "",
        scope: {
            scopeSelect: (!regions || allRegions) ? SCOPE_ITEMS.ALL.value : SCOPE_ITEMS.DEFINED.value,
            regions: REGIONS_EMPTY_VALUE,
            shouldScanStoppedInstances: shouldScanStoppedInstances || false,
            instanceTagSelector: formatTagsToStringInstances(instanceTagSelector || []),
            instanceTagExclusion: formatTagsToStringInstances(instanceTagExclusion || [])
        },
        scanFamiliesConfig: {
            sbom: {enabled: true},
            vulnerabilities: {enabled: true},
            malware: {enabled: false},
            rootkits: {enabled: false},
            secrets: {enabled: false},
            misconfigurations: {enabled: false},
            exploits: {enabled: false}
        },
        scheduled: {
            scheduledSelect: !!cronLine ? SCHEDULE_TYPES_ITEMS.REPETITIVE.value : SCHEDULE_TYPES_ITEMS.NOW.value,
            laterDate: "",
            laterTime: "",
            cronLine: cronLine || CRON_QUICK_OPTIONS[0].value
        },
        maxParallelScanners: maxParallelScanners || 2,
        scannerInstanceCreationConfig: {
            useSpotInstances: useSpotInstances || false
        }
    }
    
    if (!isEmpty(regions)) {
        initialValues.scope.regions = regions.map(({name, vpcs}) => {
            return {name, vpcs: !vpcs ? VPCS_EMPTY_VALUE : vpcs.map(({id, securityGroups}) => {
                return {id: id || "", securityGroups: (securityGroups || []).map(({id}) => id)}
            })}
        })
    }
    
    if (!!operationTime && !cronLine) {
        const dateTime = new Date(operationTime);
        initialValues.scheduled.scheduledSelect = SCHEDULE_TYPES_ITEMS.LATER.value;
        initialValues.scheduled.laterTime = `${padDateTime(dateTime.getHours())}:${padDateTime(dateTime.getMinutes())}`;
        initialValues.scheduled.laterDate = `${dateTime.getFullYear()}-${padDateTime(dateTime.getMonth() + 1)}-${padDateTime(dateTime.getDate())}`;
    }

    Object.keys(scanFamiliesConfig || {}).forEach(type => {
        const {enabled} = scanFamiliesConfig[type];
        initialValues.scanFamiliesConfig[type].enabled = enabled;
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
                const {id, scope, scheduled, ...submitData} = formValues;

                const {scopeSelect, regions, shouldScanStoppedInstances, instanceTagSelector, instanceTagExclusion} = scope;
                const isAllScope = scopeSelect === SCOPE_ITEMS.ALL.value;

                submitData.scope = {
                    objectType: "AwsScanScope",
                    allRegions: isAllScope,
                    regions: isAllScope ? [] : regions.map(({name, vpcs}) => {
                        return {name, vpcs: vpcs.map(({id, securityGroups}) => {
                            return {id, securityGroups: securityGroups.map(id => ({id}))}
                        })}
                    }),
                    shouldScanStoppedInstances,
                    instanceTagSelector: formatStringInstancesToTags(instanceTagSelector),
                    instanceTagExclusion: formatStringInstancesToTags(instanceTagExclusion),
                }

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