import React, { useEffect, useState, useMemo } from 'react';
import { cloneDeep, isNull, isEqual } from 'lodash';
import classnames from 'classnames';
import { useField } from 'formik';
import { usePrevious } from 'hooks';
import DropdownSelect from 'components/DropdownSelect';
import Loader from 'components/Loader';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';

import './select-field.scss';

const getMissingValueItemKeys = (valueKey, items) => {
    if (isNull(valueKey)) {
        return items;
    }

    const valueInItems = items.find(item => item.value === valueKey);

    if (!valueInItems) {
        items = cloneDeep(items);
        items.push({value: valueKey, label: valueKey});
    }

    return items;
}

const SelectField = (props) => {
    const {items: fieldItems=[], placeholder, creatable=false, clearable=false, disabled, className, label, tooltipText,
        components={}, loading} = props;
    const [field, meta, helpers] = useField(props);
    const {value} = meta;
    const {name} = field;
    const {setValue, setTouched} = helpers;
    
    const formattedItems = useMemo(() => (
        creatable && value !== "" ? getMissingValueItemKeys(value, fieldItems) : fieldItems
    ), [creatable, fieldItems, value]);
    const prevFormattedItems = usePrevious(formattedItems);
    
    const [items, setItems] = useState(formattedItems);

    useEffect(() => {
        if (!isEqual(formattedItems, prevFormattedItems)) {
            setItems(formattedItems);
        }
    }, [prevFormattedItems, formattedItems]);

    const selectedValue = items.find(item => item.value === value) || null;
    
    return (
        <div className={classnames("form-field-wrapper", "select-field-wrapper", className)}>
            {!!label && <FieldLabel tooltipId={`form-tooltip-${name}`} tooltipText={tooltipText}>{label}</FieldLabel>}
            <DropdownSelect
                name={name}
                value={selectedValue}
                items={items}
                onChange={selectedItem => {
                    const {value} = selectedItem || {};
                    
                    if (creatable) {
                        setItems(getMissingValueItemKeys(value, items));
                    }
                    
                    setTouched(true, true);
                    setValue(value);
                }}
                onBlur={() => setTouched(true, true)}
                creatable={creatable}
                clearable={clearable}
                disabled={disabled || loading}
                placeholder={placeholder}
                components={components}
            />
            {loading && <Loader small />}
            {meta.touched && meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default SelectField;