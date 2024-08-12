import React, { useEffect } from 'react';
import { Formik, Form, useFormikContext } from 'formik';
import { isNull, isEmpty, cloneDeep } from 'lodash';
import classnames from 'classnames';

import Loader from 'components/Loader';
import Button from 'components/Button';
import Icon, { ICON_NAMES } from 'components/Icon';
import { useFetch, FETCH_METHODS, usePrevious } from 'hooks';
import * as validators from './validators';
import SelectField from './form-fields/SelectField';
import MultiselectField from './form-fields/MultiselectField';
import TextField from './form-fields/TextField';
import TextAreaField from './form-fields/TextAreaField';
import RadioButtonGroup from './form-fields/RadioButtonGroup';
import FieldsPair from './form-fields/FieldsPair';
import CheckboxField from './form-fields/CheckboxField';
import DateField from './form-fields/DateField';
import TimeField from './form-fields/TimeField';
import CronField from './form-fields/CronField';
import FieldLabel from './FieldLabel';

import './form.scss';

export {
	CheckboxField,
	CronField,
	DateField,
	FieldLabel,
	FieldsPair,
	MultiselectField,
	RadioButtonGroup,
	SelectField,
	TextAreaField,
	TextField,
	TimeField,
	useFormikContext,
	validators,
}

const FormComponent = ({ children, className, submitUrl, getSubmitParams, onSubmitSuccess, onSubmitError, saveButtonTitle = "Finish", hideSaveButton = false }) => {
	const { values, isSubmitting, isValidating, setSubmitting, status, setStatus, isValid, setErrors } = useFormikContext();

	const [{ loading, data, error }, submitFormData] = useFetch(submitUrl, { loadOnMount: false });
	const prevLoading = usePrevious(loading);

	const handleSubmit = () => {
		setSubmitting(true);

		const submitQueryParams = !!getSubmitParams ? getSubmitParams(cloneDeep(values)) : {};
		submitFormData({ method: FETCH_METHODS.POST, submitData: values, ...submitQueryParams });
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
				const { message, errors } = error;

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
		<Form className={classnames("form-wrapper", { [className]: className })}>
			{!!status &&
				<div className="main-error-message">
					<Icon name={ICON_NAMES.ALERT} />
					<div>{status}</div>
				</div>
			}
			{children}
			{!hideSaveButton &&
				<Button type="submit" className="form-submit-button" onClick={handleSubmit} disabled={disableSubmitClick}>
					{saveButtonTitle}
				</Button>
			}
		</Form>
	)
}

const FormWrapper = ({ children, initialValues, validate, ...props }) => {
	return (
		<Formik initialValues={initialValues} validate={validate} validateOnMount={true}>
			<FormComponent {...props}>
				{children}
			</FormComponent>
		</Formik>
	)
}

export default FormWrapper;