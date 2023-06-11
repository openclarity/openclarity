import React from 'react';
import { FETCH_METHODS, useFetch } from 'hooks';
import FormWrapper, { ToggleField, TextField, validators } from 'components/Form';
import Modal from 'components/Modal';
import Loader from 'components/Loader';

const SCAN_CONFIG_URL = "runtime/quickscan/config";

const ScanOptionsForm = ({onClose}) => {
    const [{data, loading}] = useFetch(SCAN_CONFIG_URL);
    
    const initialValues = {
        cisDockerBenchmarkScanEnabled: data?.cisDockerBenchmarkScanEnabled || false,
        maxScanParallelism: data?.maxScanParallelism || 10
    };
    
    return (
        <Modal
            title="On-demand scan options"
            className="scan-options-form-modal"
            onClose={onClose}
            stickLeft
            hideCancel
            hideSubmit
            width={530}
        >
            {loading ? <Loader /> :
                <FormWrapper
                    initialValues={initialValues}
                    submitUrl={SCAN_CONFIG_URL}
                    onSubmitSuccess={onClose}
                    saveButtonTitle="Save"
                    getSubmitParams={formValues => ({
                        method: FETCH_METHODS.PUT,
                        submitData: formValues
                    })}
                >
                    <ToggleField name="cisDockerBenchmarkScanEnabled" label="CIS Docker Benchmark" />
                    <TextField 
                        name="maxScanParallelism" 
                        label="Max Scan Parallelism" 
                        type="number" 
                        validate={validators.validateMaxScanField}
                        min="1" 
                    />
                </FormWrapper>
            }
        </Modal>
    )
}

export default ScanOptionsForm;