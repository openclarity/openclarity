import React, { useEffect } from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { FieldArray, useField, useFormikContext } from 'formik';
import { usePrevious } from 'hooks';
import FieldLabel from 'components/Form/FieldLabel';
import Icon, { ICON_NAMES } from 'components/Icon';

import './fields-pair.scss';

const FieldAction = ({iconName, onClick, disabled, count}) => (
    <div className={classnames("field-action", {disabled})} onClick={disabled ? undefined : onClick}>
        <Icon name={iconName} />
    </div>
)

const FieldItemWrapper = ({name, value, index, push, remove, replace, firstFieldProps, secondFieldProps, disabled, count}) => {
    const {component: FirstFieldComponent, emptyValue: firstEmptyValue="", ...firstProps} = firstFieldProps;
    const {component: SecondFieldComponent, getDependentFieldProps, emptyValue: secondEmptyValue="", ...secondProps} = secondFieldProps;

    const firstKey = firstProps.key;
    const secondKey = secondProps.key;

    const firstValue = value[index][firstKey];
    const prevFirstValue = usePrevious(firstValue);

    const prevCount = usePrevious(count);

    const allowRemove = value.length > 1;

    const formattedFirstProps = {
        ...firstProps,
        disabled
    };
    const formattedSecondProps = {
        ...secondProps,
        ...(!getDependentFieldProps ? {} : {...getDependentFieldProps(value[index]), index})
    };
    formattedSecondProps.disabled = disabled || formattedSecondProps.disabled || isEmpty(firstValue);
    
    useEffect(() => {
        if (count === prevCount && !!prevFirstValue && prevFirstValue !== firstValue) {
            replace(index, {...value[index], [secondKey]: secondEmptyValue}); 
        }
    }, [prevFirstValue, firstValue, index, value, secondKey, secondEmptyValue, replace, count, prevCount]);

    return (
        <div key={index} className="fields-wrapper">
            <div className="field-with-actions-container">
                <FirstFieldComponent
                    index={index}
                    name={`${name}.${index}.${firstKey}`}
                    {...formattedFirstProps}
                />
                <div className="actions-wrapper">
                    <FieldAction
                        iconName={ICON_NAMES.MINUS}
                        onClick={() => remove(index)}
                        disabled={disabled || !allowRemove}
                    />
                    <FieldAction
                        iconName={ICON_NAMES.PLUS}
                        onClick={() => push({[firstKey]: firstEmptyValue, [secondKey]: secondEmptyValue})}
                        disabled={disabled}
                    />
                </div>
            </div>
            <SecondFieldComponent
                index={index}
                name={`${name}.${index}.${secondKey}`}
                {...formattedSecondProps}
            />
        </div>
    );
}

const FieldsPair = (props) => {
    const {label, className, tooltipText} = props;
    const [field] = useField(props);
    
    const {name, value} = field;

    const {validateForm} = useFormikContext();

    useEffect(() => {
        validateForm();
    }, [value, validateForm])
    
    return (
        <div className={classnames("form-field-wrapper", "fields-pair-field-wrapper", {[className]: className})}>
            {!isEmpty(label) && <FieldLabel tooltipId={name} tooltipText={tooltipText}>{label}</FieldLabel>}
            <FieldArray name={name}>
                {({remove, push, replace}) => value.map((item, index) => (
                    <FieldItemWrapper
                        key={index}
                        {...props}
                        name={name}
                        index={index}
                        value={value}
                        remove={remove}
                        push={push}
                        replace={replace}
                        count={value.length}
                    />
                ))}
            </FieldArray>
        </div>
    )
}

export default FieldsPair;