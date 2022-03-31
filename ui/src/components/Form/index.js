import React, { useEffect } from 'react';
import { Formik, Form, useFormikContext } from 'formik';
import { isNull, isEmpty, cloneDeep } from 'lodash';
import classnames from 'classnames';
import Loader from 'components/Loader';
import Button from 'components/Button';
import Icon, { ICON_NAMES } from 'components/Icon';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import SelectField from './form-fields/SelectField';
import MultiselectField from './form-fields/MultiselectField';
import TextField from './form-fields/TextField';

import './form.scss';

export {
    SelectField,
    MultiselectField,
    TextField
}

const FormComponent = ({children, className, submitUrl, getSubmitParams, onSubmitSuccess, onSubmitError, saveButtonTitle="Finish"}) => {
	const {values, isSubmitting, isValidating, setSubmitting, status, setStatus, isValid, setErrors} = useFormikContext();

	const [{loading, data, error}, submitFormData] = useFetch(submitUrl, {loadOnMount: false});
	const prevLoading = usePrevious(loading);

	const handleSubmit = () => {
		setSubmitting(true);
        
		const submitQueryParams = !!getSubmitParams ? getSubmitParams(cloneDeep(values)) : {};
		submitFormData({method: FETCH_METHODS.POST, submitData: values, ...submitQueryParams});
    }
	
	useEffect(() => {
		if (prevLoading && !loading) {
			setSubmitting(false);
			setStatus(null);
			
			if (isNull(error)) {
				if (!!onSubmitSuccess) {
					onSubmitSuccess(data);
				}
			} else {
				const {message, errors} = error;

				if (!!message) {
					setStatus(message); 
				}
				
				if (!isEmpty(errors)) {
					setErrors(errors);
				}

				if (!!onSubmitError) {
					onSubmitError();
				}
			}
		}
	}, [prevLoading, loading, error, data, setSubmitting, setStatus, onSubmitSuccess, setErrors, onSubmitError]);

	if (isSubmitting || loading) {
		return <Loader />;
	}

	const disableSubmitClick = isSubmitting || isValidating || !isValid;

	return (
		<Form className={classnames("form-wrapper", {[className]: className})}>
			{!!status &&
				<div className="main-error-message">
					<Icon name={ICON_NAMES.ALERT} />
					<div>{status}</div>
				</div>
			}
			{children}
            <Button type="submit" className="form-submit-button" onClick={handleSubmit} disabled={disableSubmitClick}>
                {saveButtonTitle}
            </Button>
		</Form>
	)
}

const FormWrapper = ({children, initialValues, validate, ...props}) => {
	return (
		<Formik initialValues={initialValues} validate={validate} validateOnMount={true}>
			<FormComponent {...props}>
				{children}
			</FormComponent>
		</Formik>
	)
}

export default FormWrapper;