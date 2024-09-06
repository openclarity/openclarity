import React, { useState, useEffect } from "react";
import classnames from "classnames";
import { Formik, Form, useFormikContext } from "formik";
import { cloneDeep } from "lodash";
import Button from "components/Button";
import Modal from "components/Modal";
import Loader from "components/Loader";
import Title from "components/Title";
import Arrow, { ARROW_NAMES } from "components/Arrow";
import { useMutation } from "@tanstack/react-query";
import { openClarityApi } from "../../api/openClarityApi";

import "./wizard-modal.scss";

function Wizard({
  steps,
  onClose,
  onSubmitSuccess,
  getSubmitParams,
  isEditForm,
}) {
  const {
    values,
    isSubmitting,
    isValidating,
    status,
    setStatus,
    isValid,
    setErrors,
    validateForm,
  } = useFormikContext();

  const [activeStepId, setActiveStepId] = useState(steps[0].id);

  const activeStepIndex = steps.findIndex(({ id }) => id === activeStepId);
  const { component: ActiveStepComponent, title: activeTitle } =
    steps[activeStepIndex];
  const { title: nextStepTitle, id: nextStepId } =
    steps[activeStepIndex + 1] || {};

  const disableStepDone = isSubmitting || isValidating || !isValid;

  const scanConfigMutation = useMutation({
    mutationFn: () => {
      const submitParams = getSubmitParams(cloneDeep(values));
      if (isEditForm) {
        return openClarityApi.putScanConfigsScanConfigID(
          values.id,
          submitParams.submitData,
        );
      }
      return openClarityApi.postScanConfigs(submitParams.submitData);
    },
    onSuccess: () => {
      onSubmitSuccess();
    },
    onError: (error) => {
      setStatus(error.message);
      setErrors(error.errors);
    },
  });

  const onStepClick = (stepId) => {
    if (disableStepDone || stepId === activeStepId) {
      return;
    }

    setActiveStepId(stepId);
  };

  useEffect(() => {
    validateForm();
  }, [activeStepId, validateForm]);

  if (scanConfigMutation.isPending) {
    return <Loader />;
  }

  return (
    <Form className="wizard-wrapper">
      <div className="wizard-content">
        <ul className="wizard-navigation">
          {steps.map(({ id, title }) => (
            <li
              key={id}
              className={classnames(
                "wizard-navigation-item",
                { "is-active": id === activeStepId },
                { disabled: disableStepDone },
              )}
              onClick={() => onStepClick(id)}
            >
              {title}
            </li>
          ))}
        </ul>
        <div className="wizard-step-display">
          {!!status && <div className="main-error-message">{status}</div>}
          <Title medium>{activeTitle}</Title>
          <ActiveStepComponent />
          {!!nextStepTitle && (
            <div
              className={classnames("wizard-next-step-wrapper", {
                disabled: disableStepDone,
              })}
              onClick={
                disableStepDone ? undefined : () => onStepClick(nextStepId)
              }
            >
              <div className="wizard-next-step-title">{`Go to ${nextStepTitle}`}</div>
              <Arrow name={ARROW_NAMES.RIGHT} />
            </div>
          )}
        </div>
      </div>
      <div className="wizard-action-buttons">
        <Button tertiary onClick={onClose}>
          Cancel
        </Button>
        <Button
          onClick={() => scanConfigMutation.mutate()}
          disabled={disableStepDone}
        >
          Save
        </Button>
      </div>
    </Form>
  );
}

const WizardModal = ({ title, initialValues, onClose, validate, ...props }) => (
  <Modal
    className="wizard-modal"
    title={title}
    onClose={onClose}
    disableDone
    hideCancel
    hideSubmit
    stickLeft
  >
    <Formik initialValues={initialValues} validate={validate}>
      <Wizard {...props} onClose={onClose} />
    </Formik>
  </Modal>
);

export default WizardModal;
