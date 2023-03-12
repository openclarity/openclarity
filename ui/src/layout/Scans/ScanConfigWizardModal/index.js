import React from 'react';
import { FETCH_METHODS } from 'hooks';
import WizardModal from 'components/WizardModal';
import { formatStringInstancesToTags, formatTagsToStringInstances } from '../utils';
import StepGeneralProperties, { REGIONS_EMPTY_VALUE, VPCS_EMPTY_VALUE, SCOPE_ITEMS } from './StepGeneralProperties';
import StepScanTypes from './StepScanTypes';
import StepTimeConfiguration, { SCHEDULE_TYPES_ITEMS } from './StepTimeConfiguration';

import './scan-config-wizard-modal.scss';

const padDateTime = time => String(time).padStart(2, "0");

const ScanConfigWizardModal = ({initialData, onClose, onSubmitSuccess}) => {
    const {id, name, scope, scanFamiliesConfig, scheduled} = initialData || {};
    const {all, regions, shouldScanStoppedInstances, instanceTagSelector, instanceTagExclusion} = scope || {};
    
    const isEditForm = !!id;
    
    const initialValues = {
        id: id || null,
        name: name || "",
        scope: {
            scopeSelect: (!regions || all) ? SCOPE_ITEMS.ALL.value : SCOPE_ITEMS.DEFINED.value,
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
            scheduledSelect: !!scheduled?.objectType ? SCHEDULE_TYPES_ITEMS.LATER.value : SCHEDULE_TYPES_ITEMS.NOW.value,
            laterDate: "",
            laterTime: ""
        }
    }
    
    if (!!regions) {
        initialValues.scope.regions = regions.map(({id, vpcs}) => {
            return {id, vpcs: !vpcs ? VPCS_EMPTY_VALUE : vpcs.map(({id, securityGroups}) => {
                return {id: id || "", securityGroups: (securityGroups || []).map(({id}) => id)}
            })}
        })
    }
    
    const {operationTime} = scheduled || {};
    if (!!operationTime) {
        const dateTime = new Date(operationTime);
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
        }
    ];

    return (
        <WizardModal
            title={`${isEditForm ? "Edit" : "New"} scan config`}
            onClose={onClose}
            steps={steps}
            initialValues={initialValues}
            submitUrl="scanConfigs"
            getSubmitParams={formValues => {
                const {id, scope, scheduled, ...submitData} = formValues;

                const {scopeSelect, regions, shouldScanStoppedInstances, instanceTagSelector, instanceTagExclusion} = scope;
                const isAllScope = scopeSelect === SCOPE_ITEMS.ALL.value;

                submitData.scope = {
                    objectType: "AwsScanScope",
                    all: isAllScope,
                    regions: isAllScope ? null : regions.map(({id, vpcs}) => {
                        return {id, vpcs: vpcs.map(({id, securityGroups}) => {
                            return {id, securityGroups: securityGroups.map(id => ({id}))}
                        })}
                    }),
                    shouldScanStoppedInstances,
                    instanceTagSelector: formatStringInstancesToTags(instanceTagSelector),
                    instanceTagExclusion: formatStringInstancesToTags(instanceTagExclusion),
                }

                regions.map(({id, vpcs}) => {
                    return {id, vpcs: !vpcs ? VPCS_EMPTY_VALUE : vpcs.map(({id, securityGroups}) => {
                        return {id: id || "", securityGroups: (securityGroups || []).map(({id}) => id)}
                    })}
                })

                const {scheduledSelect, laterDate, laterTime} = scheduled;
                const isNow = scheduledSelect === SCHEDULE_TYPES_ITEMS.NOW.value;
                
                let formattedDate = new Date();

                if (!isNow) {
                    const [hours, minutes] = laterTime.split(":");
                    formattedDate = new Date(laterDate);
                    formattedDate.setHours(hours, minutes);
                }

                submitData.scheduled = {
                    objectType: "SingleScheduleScanConfig",
                    operationTime: formattedDate.toISOString()
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