import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFetch, usePrevious } from 'hooks';
import { components } from 'react-select';
import { useFilterDispatch, setFilters, FILTERR_TYPES } from 'context/FiltersProvider';
import Icon, { ICON_NAMES } from 'components/Icon';
import { OPERATORS } from 'components/Filter';
import Button from 'components/Button';
import Loader from 'components/Loader';
import FormWrapper, { SelectField, useFormikContext } from 'components/Form';
import { CIS_SEVERITY_ITEMS, ROUTES, SEVERITY_ITEMS } from 'utils/systemConsts';
import StepDisplay from '../StepDisplay';

import './severity-filter-and-counters-steps.scss';

const CIS_LINKS_DISPLAY_MAP = [
    {title: "Applications", dataKey: "applications", route: ROUTES.APPLICATIONS, filter: FILTERR_TYPES.APPLICATIONS},
    {title: "Application Resources", dataKey: "resources", route: ROUTES.APPLICATION_RESOURCES, filter: FILTERR_TYPES.APPLICATION_RESOURCES}
];

const LINKS_DISPLAY_MAP = [
    ...CIS_LINKS_DISPLAY_MAP,
    {title: "Packages", dataKey: "packages", route: ROUTES.PACKAGES, filter: FILTERR_TYPES.PACKAGES},
    {title: "Vulnerabilities", dataKey: "vulnerabilities", route: ROUTES.VULNERABILITIES, filter: FILTERR_TYPES.VULNERABILITIES}
];

const FILTER_TYPE_ITEMS = {
    VULNERABILITY: {value: "VULNERABILITY", label: "Vulnerability", filterName: "vulnerabilitySeverity[gte]"},
    CIS: {value: "CIS", label: "CIS Benchmark", filterName: "cisDockerBenchmarkLevel[gte]"}
}

const SelectItem = ({label, color, icon}) => (
    <div style={{display: "flex", alignItems: "center"}}>
        <Icon name={icon} style={{color}} />
        <div style={{marginLeft: "5px"}}>{label}</div>
    </div>
)

const isVulnerabilityType = type => type === FILTER_TYPE_ITEMS.VULNERABILITY.value;
const getSeverityItemsByType = type => isVulnerabilityType(type) ?  SEVERITY_ITEMS : CIS_SEVERITY_ITEMS;

const FormFields = ({onChange}) => {
    const {values, setFieldValue} = useFormikContext();
    const {type, severity} = values;
    const prevType = usePrevious(type);
    const prevSeverity = usePrevious(severity);

    const icon = isVulnerabilityType(type) ? ICON_NAMES.BUG : ICON_NAMES.ALERT;

    useEffect(() => {
        if (prevType !== type) {
            setFieldValue("severity", Object.keys(getSeverityItemsByType(type)).at(-1));
        }
    }, [prevType, type, setFieldValue]);

    useEffect(() => {
        if (prevSeverity !== severity) {
            onChange({type, severity});
        }
    }, [prevSeverity, severity, type, onChange]);

    return (
        <React.Fragment>
            <SelectField
                name="type"
                items={Object.values(FILTER_TYPE_ITEMS)}
            />
            <SelectField
                name="severity"
                items={Object.values(getSeverityItemsByType(type))}
                components={{
                    Option: props => {
                        const {label, color} = props.data;

                        return (
                            <components.Option {...props}>
                                <SelectItem label={label} color={color} icon={icon} />
                            </components.Option>
                        )
                    },
                    SingleValue: props => {
                        const {label, color} = props.data;

                        return (
                            <components.SingleValue {...props}>
                                <SelectItem label={label} color={color} icon={icon} />
                            </components.SingleValue>
                        )
                    }
                }}
            />
        </React.Fragment>
    )
}

const FilterForm = ({onChange}) => {
    return (
        <FormWrapper
            className="severity-filter-form"
            initialValues={{type: FILTER_TYPE_ITEMS.VULNERABILITY.value, severity: SEVERITY_ITEMS.NEGLIGIBLE.value}}
            hideSaveButton
        >
            <FormFields onChange={onChange} />
        </FormWrapper>
    )
}

const SeverityFilterAndCountersSteps = () => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const [{loading, data}, fetchData] = useFetch("runtime/scan/results", {loadOnMount: false});

    const [selectedFilter, setSelectedFilter] = useState(null);
    const {severity, type} = selectedFilter || {};
    const activeFilter = FILTER_TYPE_ITEMS[type]?.filterName;

    useEffect(() => {
        if (!!severity) {
            fetchData({queryParams: {[activeFilter]: severity}});
        }
    }, [severity, activeFilter, fetchData]);

    const {counters={}} = data || {};
    
    return (
        <React.Fragment>
            <StepDisplay step="3"  title="Filter:" className="filter-step">
                <FilterForm onChange={setSelectedFilter} />
            </StepDisplay>
            <StepDisplay step="4"  title="Affected elements:" className="affected-elements-step">
                    <div className="select-counter-buttons-wrapper">
                        {
                            (isVulnerabilityType(type) ? LINKS_DISPLAY_MAP : CIS_LINKS_DISPLAY_MAP).map((linkItem) => {
                                const {title, dataKey} = linkItem;
                                const itemCount = counters[dataKey] || 0;

                                return {...linkItem, title: `${title} (${itemCount})`, disabled: !itemCount};
                            }).map(({title, dataKey, route, filter, disabled}, index, items) => {
                                const onCounterClick = () => {
                                    setFilters(filtersDispatch, {type: filter, filters: {currentRuntimeScan: {items, current: dataKey, minimalSeverity: severity}}, isSystem: true});
                                    setFilters(filtersDispatch, {type: filter, filters: [{scope: "vulnerabilitySeverity", operator: OPERATORS.gte.value, value: [severity]}], isSystem: false});
                                    navigate(route);
                                }
            
                                return (
                                    <Button key={dataKey} disabled={disabled} onClick={onCounterClick}>{title}</Button>
                                )
                            })
                        }
                    </div>
                    {/* <div style={{position: "relative"}}><Loader small /></div> */}
                    {/* {loading && <div style={{position: "relative"}}><Loader small /></div>} */}
            </StepDisplay>
        </React.Fragment>
    )
}

export default SeverityFilterAndCountersSteps;