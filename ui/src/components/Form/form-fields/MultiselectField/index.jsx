import React, { useState, useEffect, useMemo } from 'react';
import { cloneDeep, isEqual, isEmpty } from 'lodash';
import classnames from 'classnames';
import { useField } from 'formik';
import { components } from 'react-select';
import DropdownSelect from 'components/DropdownSelect';
import Loader from 'components/Loader';
import FieldError from 'components/Form/FieldError';
import FieldLabel from 'components/Form/FieldLabel';
import { usePrevious } from 'hooks';

import './multiselect-field.scss';

const ConnectorMultiValueContainer = ({connector, ...props}) => (
    <div className="multi-select-custom-item-with-connector">
        <div className="multi-select-connector">{connector}</div>
        <components.MultiValueContainer {...props} />
    </div>
)

const PrevixLabelControl = ({children, label, ...props}) => {
    const hasValue = !isEmpty(props.getValue());
    
    return (
        <components.Control {...props} className="multi-select-custom-control-with-label">
            {hasValue && <span className="multi-select-control-label">{label}</span>}
            {children}
        </components.Control>
    )
}

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
    const {items: fieldItems=[], placeholder, creatable=false, disabled, className, label, tooltipText, connector,
        prefixLabel, loading} = props;
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

    const selectedItems = items.filter(item => value?.includes(item.value));
    
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
                disabled={disabled || loading}
                placeholder={placeholder}
                isMulti={true}
                components={{
                    ...(connector ? {MultiValueContainer: props => <ConnectorMultiValueContainer connector={connector} {...props} />} : {}),
                    ...(prefixLabel ? {Control: props => <PrevixLabelControl label={prefixLabel} {...props} />} : {}),
                }}
            />
            {loading && <Loader small />}
            {meta.error && <FieldError>{meta.error}</FieldError>}
        </div>
    )
}

export default MultiselectField;