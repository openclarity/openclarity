import React from 'react';
import { FETCH_METHODS, useFetch } from 'hooks';
import FormWrapper, { ToggleField } from 'components/Form';
import Modal from 'components/Modal';
import Loader from 'components/Loader';

const SCAN_CONFIG_URL = "runtime/quickscan/config";

const ScanOptionsForm = ({onClose}) => {
    const [{data, loading}] = useFetch(SCAN_CONFIG_URL);
    
    const initialValues = {
        cisDockerBenchmarkScanEnabled: data?.cisDockerBenchmarkScanEnabled || false
    };
    
    return (
        <Modal
            title="On-demand scan options"
            className="scan-options-form-modal"
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
                    getSubmitParams={formValues => ({
                        method: FETCH_METHODS.PUT,
                        submitData: formValues
                    })}
                >
                    <ToggleField name="cisDockerBenchmarkScanEnabled" label="CIS Docker Benchmark" />
                </FormWrapper>
            }
        </Modal>
    )
}

export default ScanOptionsForm;