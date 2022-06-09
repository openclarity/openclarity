import React from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/creatable';
import classnames from 'classnames';

import COLORS from 'utils/scss_variables.module.scss';

import './dropdown-select.scss';

const DropdownSelect = ({items, value, onChange, creatable=false, clearable=false, disabled=false, placeholder="Select...", isMulti=false, className, components, small}) => {
    const SelectComponent = creatable ? CreatableSelect : Select;
    const height = small ? 20 : 36;
    const innerHeight = height - 2;
    
    return (
        <SelectComponent
            value={value}
            onChange={onChange}
            className={classnames("dropdown-select", {small}, className)}
            classNamePrefix="dropdown-select"
            options={items}
            isClearable={clearable}
            isDisabled={disabled}
            placeholder={placeholder}
            isMulti={isMulti}
            components={components}
            styles={{
                control: (provided) => ({
                    ...provided,
                    minHeight: height,
                    borderRadius: 2,
                    borderColor: COLORS["color-grey-light"],
                    boxShadow: "none",
                    "&:hover": {
                        ...provided["&:hover"],
                        borderColor: COLORS["color-grey-light"]
                    },
                    backgroundColor: "white",
                    cursor: "pointer",
                    fontSize: small ? 10 : 14,
                    lineHeight: small ? "14px" : "18px"
                }),
                option: (provided, state) => {
                    const {isSelected, isDisabled} = state;
                    
                    return ({
                        ...provided,
                        color: isSelected ? COLORS["color-grey-dark"] : (isDisabled ? COLORS["color-grey-light"] : COLORS["color-grey-dark"]),
                        backgroundColor: isSelected ? COLORS["color-grey-lighter"] : "transparent",
                        fontWeight: isSelected ? "bold" : "normal",
                        cursor: "pointer"
                    });
                },
                placeholder: (provided, state) => ({
                    ...provided,
                    color: state.isDisabled ? COLORS["color-grey"] : COLORS["color-grey-dark"],
                    ...((small && !isMulti) ? {height: innerHeight} : {})
                }),
                menu: (provided) => ({
                    ...provided,
                    borderRadius: 2,
                    border: `1px solid ${COLORS["color-grey"]}`,
                    borderTop: `2px solid ${COLORS["color-main-light"]}`,
                    fontSize: 14,
                    lineHeight: "18px"
                }),
                indicatorsContainer: (provided) => ({
                    ...provided,
                    height: innerHeight
                }),
                indicatorSeparator: (provided) => ({
                    ...provided,
                    display: small ? "none" : "block"
                }),
                dropdownIndicator: (provided) => ({
                    ...provided,
                    padding: small ? 0 : 8,
                }),
                valueContainer: (provided) => ({
                    ...provided,
                    minHeight: innerHeight,
                    padding: (small && isMulti) ? "0 8px" : "2px 8px"
                }),
                singleValue: (provided) => ({
                    ...provided,
                    ...(small ? {height: innerHeight} : {})
                }),
                multiValue: (provided) => ({
                    ...provided,
                    borderRadius: "100px",
                    ...(small ? {height: 14} : {})
                }),
                multiValueLabel: (provided, state) => {
                    const {isDisabled} = state;

                    return ({
                        ...provided,
                        color: isDisabled ? COLORS["color-grey"] : COLORS["color-main"],
                        backgroundColor: isDisabled ? COLORS["color-grey-lighter"] : COLORS["color-blue"],
                        borderRadius: "100px 0 0 100px",
                        border: `1px solid ${isDisabled ? COLORS["color-grey"] : COLORS["color-main-light"]}`,
                        borderRight: "none",
                        padding: small ? "0 3px" : "3px"
                    })
                },
                multiValueRemove: (provided, state) => {
                    const {isDisabled} = state;
                    const backgroundColor = isDisabled ? COLORS["color-grey-lighter"] : COLORS["color-blue"];
                    const color = isDisabled ? COLORS["color-grey"] : COLORS["color-main"];

                    return ({
                        ...provided,
                        ":hover": {
                            ...provided[":hover"],
                            color: color,
                            backgroundColor: backgroundColor,
                            cursor: "pointer"
                        },
                        color: color,
                        backgroundColor: backgroundColor,
                        borderRadius: "0 100px 100px 0",
                        border: `1px solid ${isDisabled ? COLORS["color-grey"] : COLORS["color-main-light"]}`,
                        borderLeft: "none"
                    })
                }
            }}
        />
    )
}

export default DropdownSelect;