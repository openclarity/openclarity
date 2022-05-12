import React, { useState, useEffect, useMemo } from 'react';
import { cloneDeep, isEqual } from 'lodash';
import classnames from 'classnames';
import { useField } from 'formik';
import DropdownSelect from 'components/DropdownSelect';
import { FieldLabel, FieldError } from 'components/Form/utils';
import { usePrevious } from 'hooks';

const getMissingValueItemKeys = (valueKeys, items) => {
    const missingItems = valueKeys.filter(key => !items.find(item => item.value === key));

    if(missingItems.length > 0) {
        items = cloneDeep(items);
        missingItems.forEach(item => {
            items.push({value: item, label: item});
        });
    }

    return items;
}

const MultiselectField = (props) => {
    const {items: fieldItems=[], placeholder, creatable=false, disabled, className, small, label, tooltipText} = props;
    const [field, meta, helpers] = useField(props);
    const {value} = meta;
    const {name} = field;
    const {setValue} = helpers;
    
    const formattedItems = useMemo(() => (
        creatable ? getMissingValueItemKeys(value || [], fieldItems) : fieldItems
    ), [creatable, fieldItems, value]);
    const prevFormattedItems = usePrevious(formattedItems);

    const [items, setItems] = useState(formattedItems);

    useEffect(() => {
        if (!isEqual(formattedItems, prevFormattedItems)) {
            setItems(formattedItems);
        }
    }, [prevFormattedItems, formattedItems]);

    const selectedItems = items.filter(item => value.includes(item.value));
    
    return (
        <div className={classnames("form-field-wrapper", "multiselect-field-wrapper", className)}>
            {!!label && <FieldLabel tooltipId={`form-tooltip-${name}`} tooltipText={tooltipText}>{label}</FieldLabel>}
            <DropdownSelect
                name={name}
                value={selectedItems}
                items={items}
                onChange={selectedItems => {
                    const formattedSelectedItems = selectedItems || [];
                    const valueKeys = formattedSelectedItems.map(item => item.value);

                    if (creatable) {
                        setItems(getMissingValueItemKeys(valueKeys, items));
                    }
                    
                    setValue(valueKeys);
                }}
                creatable={creatable}
                disabled={disabled}
                placeholder={placeholder}
                isMulti={true}
                small={small}
            />
            {meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default MultiselectField;