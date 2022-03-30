import React from 'react';
import { isEmpty, isNumber, isUndefined } from 'lodash';
import { FETCH_METHODS } from 'hooks';
import FormWrapper, { SelectField, MultiselectField, TextField } from 'components/Form';
import Modal from 'components/Modal';
import { APPLICATION_TYPE_ITEMS } from './utils';

export const APP_FIELD_NAMES = {
    NAME: "name",
    TYPE: "type",
    LABELS: "labels",
    ENVIRONMENTS: "environments"
}

const validateRequired = value => (
    isEmpty(value) && !isNumber(value) && value !== 0 ? "This field is required" : undefined
);

const validateLabel = labels => {
    let hasError = false;

    labels.forEach(label => {
        const labelitems = label.split("=");
        
        if (labelitems.length !== 2 || !isUndefined(labelitems.find(item => !item))) {
            hasError = true;
        }
    });
    
    return hasError ? "Labels must be in key=value format" : undefined;
}

const FormFields = () => (
    <React.Fragment>
        <TextField name={APP_FIELD_NAMES.NAME} label="Application name" validate={validateRequired} />
        <SelectField name={APP_FIELD_NAMES.TYPE} label="Type" items={APPLICATION_TYPE_ITEMS} validate={validateRequired} />
        <MultiselectField
            name={APP_FIELD_NAMES.LABELS}
            label="Labels"
            items={[]}
            creatable
            validate={validateLabel}
            placeholder="key=value"
            tooltipText={(
                <div>
                    <div>{`Labels must be in a format of <key>=<value> where <key> and <value>`}</div>
                    <div>must consist of alphanumeric characters and also dashes (-),</div>
                    <div>underscores (_), dots (.) and slashes (/) in-between.</div>
                    <div>{`The length of <key> and <value> must be 63 characters or less.`}</div>
                    <div>The keys of the labels must be unique. Press Enter to add each label.</div>
                </div>
            )}
        />
        <MultiselectField name={APP_FIELD_NAMES.ENVIRONMENTS} label="Environments" items={[]} creatable />
    </React.Fragment>
)

const ApplicationForm = ({initialData={}, onClose, onSuccess}) => {
    const isEditForm = !!initialData.id;

    const initialValues = {
        [APP_FIELD_NAMES.NAME]: "",
        [APP_FIELD_NAMES.TYPE]: "",
        [APP_FIELD_NAMES.LABELS]: [],
        [APP_FIELD_NAMES.ENVIRONMENTS]: [],
        ...initialData
    }

    return (
        <Modal
            title={`${isEditForm ? "Edit" : "New"} Application`}
            className="application-form-modal"
            onClose={onClose}
            doneTitle={isEditForm ? "Finish" : "Create Application"}
            stickLeft
            hideCancel
            hideSubmit
        >
            <FormWrapper
                initialValues={initialValues}
                submitUrl="applications"
                getSubmitParams={formValues => {
                    const {id, ...submitData} = formValues;

                    return !isEditForm ? {submitData} : {
                        method: FETCH_METHODS.PUT,
                        formatUrl: url => `${url}/${id}`,
                        submitData
                    }
                }}
                onSubmitSuccess={onSuccess}
            >
                <FormFields />
            </FormWrapper>
        </Modal>
    )
}

export default ApplicationForm;