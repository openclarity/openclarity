import React, { useEffect, useState } from 'react';
import { isUndefined, isEmpty, isEqual } from 'lodash';
import { Formik, Form, useFormikContext } from 'formik';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import Button from 'components/Button';
import { SelectField, MultiselectField, TextField, DateField } from 'components/Form';
import CopyButton from 'components/CopyButton';
import { OPERATORS, formatFiltersToOdataItems, getValueLabel } from './utils';
import FilterButton from './FilterButton';

import './filter.scss';

export {
    OPERATORS,
    formatFiltersToOdataItems
}

const getScopeConfig = (filtersConfig, scope) => filtersConfig.find(item => item.value === scope) || {};

const DateTimeField = props => <DateField {...props} displayFormat="y-MM-dd HH:mm:ss" valueFormat="YYYY-MM-DD HH:mm:ss" />

const FormFields = ({onAdd, filtersConfig}) => {
    const {values: formValues, setFieldValue, resetForm} = useFormikContext();
    const {scope, operator, value} = formValues;

    const selectedScopeData = getScopeConfig(filtersConfig, scope);
    const {isNumber, isDate, customOdataFormat} = selectedScopeData;
    const operatorByScopeItems = selectedScopeData.operators || [];
    const selectedOperatorData = operatorByScopeItems.find(item => item.value === operator);
    const {valueItems, creatable, isSingleSelect} = selectedOperatorData || {};
    
    const ValueField = isDate ? DateTimeField : (isUndefined(valueItems) ? TextField : (isSingleSelect ? SelectField : MultiselectField));
    const valuePlaceholder = isUndefined(valueItems) ? "Enter value..." : "Select value...";
    const disableButton = isEmpty(value);

    useEffect(() => {
        setFieldValue("operator", "");
        setFieldValue("value", "");
    }, [scope, setFieldValue]);

    useEffect(() => {
        setFieldValue("value", "");
    }, [operator, setFieldValue]);

    return (
        <React.Fragment>
            <SelectField
                name="scope"
                items={filtersConfig}
                placeholder="Select property..."
            />
            <SelectField
                name="operator"
                items={operatorByScopeItems}
                placeholder="Select operator..."
                disabled={!scope}
            />
            <ValueField
                className="filter-field-value"
                name="value"
                items={valueItems}
                placeholder={valuePlaceholder}
                creatable={creatable}
                disabled={!operator}
            />
            <Button className="add-filter-button" disabled={disableButton} onClick={() => {
                if (disableButton) {
                    return;
                }
                
                onAdd({...formValues, isNumber, isDate, customOdataFormat});
                resetForm();
            }}>OK</Button>
        </React.Fragment>
    )
}

const Filter = ({filters, onFilterUpdate, filtersConfig, filtersOnCopyText}) => {
    const [showFiltersForm, setShowFiltersForm] = useState(false);

    return (
        <div className="general-filter-wrapper">
            <FilterButton onClick={() => setShowFiltersForm(!showFiltersForm)} pressed={showFiltersForm}>Filters</FilterButton>
            {showFiltersForm &&
                <div className="filter-form-wrapper">
                    <Formik
                        initialValues={{
                            scope: "",
                            operator: "",
                            value: ""
                        }}
                    >
                        <Form className="filter-form">
                            <FormFields
                                onAdd={filterData => onFilterUpdate([...filters, filterData])}
                                filtersConfig={filtersConfig}
                            />
                        </Form>
                    </Formik>
                </div>
            }
            <div className={classnames("filters-display-wrapper", {"is-empty": isEmpty(filters)})}>
                {
                    filters.map(({scope, operator, value}, index) => {
                        const {label: scopeLabel, operators: configOperators} = getScopeConfig(filtersConfig, scope);
                        
                        const operatorLabel = OPERATORS[operator].label;
                        const valueItems = configOperators.find(configOperator => configOperator.value === operator)?.valueItems || [];
                        const formattedValue = Array.isArray(value) ?
                            value.map(valueItem => getValueLabel(valueItems, valueItem)).join(" or ") : getValueLabel(valueItems, value);

                        return (
                            <div className="filter-display-item" key={index}>
                                <span>{`${scopeLabel} ${operatorLabel} `}<span style={{fontWeight: "bold"}}>{formattedValue}</span></span>
                                <Icon
                                    name={ICON_NAMES.X_MARK}
                                    onClick={() => {
                                        const newFilters = filters.filter(filterItem => !(filterItem.scope === scope && filterItem.operator === operator && isEqual(filterItem.value, value)));
                                        
                                        onFilterUpdate(newFilters); 
                                    }}
                                    size={10}
                                />
                            </div>
                        )
                    })
                }
                {!isEmpty(filters) && 
                    <>
                        <Button tertiary onClick={() => onFilterUpdate([])} >Delete all filters</Button>
                        {!!filtersOnCopyText && <div style={{marginLeft: "10px"}}><CopyButton copyText={filtersOnCopyText} /></div>}
                    </>
                }
            </div>
        </div>
    );
}

export default Filter;