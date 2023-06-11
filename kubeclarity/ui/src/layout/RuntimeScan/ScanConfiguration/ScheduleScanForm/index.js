import React, { useEffect } from 'react';
import { FETCH_METHODS, useFetch, usePrevious } from 'hooks';
import FormWrapper, { ToggleField, TextField, MultiselectField, SelectField, useFormikContext, validators } from 'components/Form';
import Modal from 'components/Modal';
import Loader from 'components/Loader';
import { NAMESPACES_SELECT_TITLE, NAMESPACES_SELECT_INFO_MESSAGE } from '../utils';
import RepetitiveFormFields, { convertServerToFormData as repetitiveConvertServerToFormData, convertFormDataToServer as repetitiveConvertFormDataToServer } from './RepetitiveFormFields';
import LaterFormFields, { convertServerToFormData as laterConvertServerToFormData, convertFormDataToServer as laterConvertFormDataToServer } from './LaterFormFields';
import { SCHEDULE_TYPES, GENERAL_FOMR_FIELDS, SCHEDULE_TYPE_DATA_WRAPPER } from './utils';

import './schedule-scan-form.scss';

const SCAN_CONFIG_URL = "runtime/scheduleScan/config";

const FormFields = ({namespaces}) => {
    const {values, setFieldValue} = useFormikContext();
    const scanConfigTypeData = values[SCHEDULE_TYPE_DATA_WRAPPER];
    const scheduleType = scanConfigTypeData[GENERAL_FOMR_FIELDS.SCHEDULE_TYPE];
    const prevScheduleType = usePrevious(scheduleType);
    
    useEffect(() => {
        if (!!prevScheduleType && prevScheduleType !== scheduleType) {
            const dataConverted = scheduleType === SCHEDULE_TYPES.LATER.value ? laterConvertServerToFormData : repetitiveConvertServerToFormData;
           setFieldValue(SCHEDULE_TYPE_DATA_WRAPPER, dataConverted(scanConfigTypeData));
        }
    }, [prevScheduleType, scheduleType, scanConfigTypeData, setFieldValue]);
    
    return (
        <React.Fragment>
            <MultiselectField
                name={GENERAL_FOMR_FIELDS.NAMESPACES}
                label={NAMESPACES_SELECT_TITLE}
                placeholder={NAMESPACES_SELECT_INFO_MESSAGE}
                tooltipText={NAMESPACES_SELECT_INFO_MESSAGE}
                items={namespaces}
            />
            <ToggleField name={GENERAL_FOMR_FIELDS.CIS_ENABLED} label="CIS Docker Benchmark" />
            <TextField 
                name={GENERAL_FOMR_FIELDS.MAX_SCANPARALLELISM} 
                label="Max Scan Parallelism" 
                type="number" 
                validate={validators.validateMaxScanField} 
                min="1" 
            />
            <SelectField
                name={`${SCHEDULE_TYPE_DATA_WRAPPER}.${GENERAL_FOMR_FIELDS.SCHEDULE_TYPE}`}
                label="Scan time"
                items={Object.values(SCHEDULE_TYPES)}
                validate={validators.validateRequired}
            />
            {values.scanConfigType.ScheduleScanConfigType === SCHEDULE_TYPES.LATER.value ? <LaterFormFields /> : <RepetitiveFormFields />}
        </React.Fragment>
    )
}

const ScheduleScanForm = ({namespaces, onClose}) => {
    const [{data, loading}] = useFetch(SCAN_CONFIG_URL);
    
    const initialValues = {
        [GENERAL_FOMR_FIELDS.NAMESPACES]: [],
        [GENERAL_FOMR_FIELDS.CIS_ENABLED]: false,
        [GENERAL_FOMR_FIELDS.MAX_SCANPARALLELISM]: 10,
        [SCHEDULE_TYPE_DATA_WRAPPER]: {
            [GENERAL_FOMR_FIELDS.SCHEDULE_TYPE]: SCHEDULE_TYPES.LATER.value
        },
        ...(data || {})
    };

    initialValues[GENERAL_FOMR_FIELDS.NAMESPACES] = initialValues[GENERAL_FOMR_FIELDS.NAMESPACES] || [];

    const laterFormData = laterConvertServerToFormData(initialValues[SCHEDULE_TYPE_DATA_WRAPPER]);
    const repetitiveFormData = repetitiveConvertServerToFormData(initialValues[SCHEDULE_TYPE_DATA_WRAPPER]);
    const selectedFormData = initialValues[SCHEDULE_TYPE_DATA_WRAPPER][GENERAL_FOMR_FIELDS.SCHEDULE_TYPE] === SCHEDULE_TYPES.LATER.value ? laterFormData : repetitiveFormData;
    initialValues[SCHEDULE_TYPE_DATA_WRAPPER] = {
        ...laterFormData,
        ...repetitiveFormData,
        [GENERAL_FOMR_FIELDS.SCHEDULE_TYPE]: selectedFormData[GENERAL_FOMR_FIELDS.SCHEDULE_TYPE]
    };
    
    return (
        <Modal
            title="Scheduled scan config"
            className="scheduled-scan-options-form-modal"
            onClose={onClose}
            stickLeft
            hideCancel
            hideSubmit
        >
            {loading ? <Loader /> :
                <FormWrapper
                    initialValues={initialValues}
                    submitUrl={SCAN_CONFIG_URL}
                    onSubmitSuccess={onClose}
                    saveButtonTitle="Save"
                    getSubmitParams={formValues => {
                        const submitData = {
                            namespaces: formValues[GENERAL_FOMR_FIELDS.NAMESPACES],
                            cisDockerBenchmarkScanEnabled: formValues[GENERAL_FOMR_FIELDS.CIS_ENABLED],
                            maxScanParallelism: formValues[GENERAL_FOMR_FIELDS.MAX_SCANPARALLELISM]
                        }

                        const scanConfigTypeData = formValues[SCHEDULE_TYPE_DATA_WRAPPER];

                        const dataConverted = scanConfigTypeData[GENERAL_FOMR_FIELDS.SCHEDULE_TYPE] === SCHEDULE_TYPES.LATER.value ? laterConvertFormDataToServer : repetitiveConvertFormDataToServer;
                        submitData.scanConfigType = dataConverted(scanConfigTypeData);
                        
                        return {
                            method: FETCH_METHODS.PUT,
                            submitData
                        } 
                    }}
                >
                    <FormFields namespaces={namespaces} />
                </FormWrapper>
            }
        </Modal>
    )
}

export default ScheduleScanForm;