import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFetch, usePrevious } from 'hooks';
import { components } from 'react-select';
import { useFilterDispatch, setFilters, setRuntimeScanFilter, FILTER_TYPES } from 'context/FiltersProvider';
import Icon, { ICON_NAMES } from 'components/Icon';
import { OPERATORS } from 'components/Filter';
import Button from 'components/Button';
import Loader from 'components/Loader';
import FormWrapper, { SelectField, useFormikContext } from 'components/Form';
import { CIS_SEVERITY_ITEMS, ROUTES, SEVERITY_ITEMS } from 'utils/systemConsts';
import StepDisplay from '../StepDisplay';

import './severity-filter-and-counters-steps.scss';

const CIS_LINKS_DISPLAY_MAP = [
    {title: "Applications", dataKey: "applications", route: ROUTES.APPLICATIONS, filter: FILTER_TYPES.APPLICATIONS},
    {title: "Application Resources", dataKey: "resources", route: ROUTES.APPLICATION_RESOURCES, filter: FILTER_TYPES.APPLICATION_RESOURCES}
];

const LINKS_DISPLAY_MAP = [
    ...CIS_LINKS_DISPLAY_MAP,
    {title: "Packages", dataKey: "packages", route: ROUTES.PACKAGES, filter: FILTER_TYPES.PACKAGES},
    {title: "Vulnerabilities", dataKey: "vulnerabilities", route: ROUTES.VULNERABILITIES, filter: FILTER_TYPES.VULNERABILITIES}
];

const FILTER_TYPE_ITEMS = {
    VULNERABILITY: {value: "VULNERABILITY", label: "Vulnerability", runtimeFilter: "vulnerabilitySeverity[gte]", tablesFilter: "vulnerabilitySeverity", countersKey: "counters"},
    CIS: {value: "CIS", label: "CIS Docker Benchmark", runtimeFilter: "cisDockerBenchmarkLevel[gte]", tablesFilter: "cisDockerBenchmarkLevel", countersKey: "cisDockerBenchmarkCounters"}
}

const SelectItem = ({label, color, icon}) => (
    <div style={{display: "flex", alignItems: "center"}}>
        <Icon name={icon} style={{color}} />
        <div style={{marginLeft: "5px"}}>{label}</div>
    </div>
)

const isVulnerabilityType = type => type === FILTER_TYPE_ITEMS.VULNERABILITY.value;
const getSeverityItemsByType = type => isVulnerabilityType(type) ?  SEVERITY_ITEMS : CIS_SEVERITY_ITEMS;

const FormFields = ({onChange, cisDockerBenchmarkScanEnabled}) => {
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
                items={[
                    FILTER_TYPE_ITEMS.VULNERABILITY,
                    ...(cisDockerBenchmarkScanEnabled ? [FILTER_TYPE_ITEMS.CIS] : [])
                ]}
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

const FilterForm = ({onChange, cisDockerBenchmarkScanEnabled}) => {
    return (
        <FormWrapper
            className="severity-filter-form"
            initialValues={{type: FILTER_TYPE_ITEMS.VULNERABILITY.value, severity: SEVERITY_ITEMS.NEGLIGIBLE.value}}
            hideSaveButton
        >
            <FormFields onChange={onChange} cisDockerBenchmarkScanEnabled={cisDockerBenchmarkScanEnabled} />
        </FormWrapper>
    )
}

const SeverityFilterAndCountersSteps = ({cisDockerBenchmarkScanEnabled}) => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const [{loading, data}, fetchData] = useFetch("runtime/scan/results", {loadOnMount: false});

    const [selectedFilter, setSelectedFilter] = useState(null);
    const {severity, type} = selectedFilter || {};
    const {runtimeFilter, tablesFilter, countersKey} = FILTER_TYPE_ITEMS[type] || {};

    useEffect(() => {
        if (!!severity) {
            fetchData({queryParams: {[runtimeFilter]: severity}});
        }
    }, [severity, runtimeFilter, fetchData]);

    const countersData = !!data ? data[countersKey] : {};
    
    return (
        <React.Fragment>
            <StepDisplay step="3"  title="Filter:" className="filter-step">
                <FilterForm onChange={setSelectedFilter} cisDockerBenchmarkScanEnabled={cisDockerBenchmarkScanEnabled} />
            </StepDisplay>
            <StepDisplay step="4"  title="Affected elements:" className="affected-elements-step">
                    <div className="select-counter-buttons-wrapper">
                        {
                            (isVulnerabilityType(type) ? LINKS_DISPLAY_MAP : CIS_LINKS_DISPLAY_MAP).map((linkItem) => {
                                const {title, dataKey} = linkItem;
                                const itemCount = countersData[dataKey] || 0;

                                return {...linkItem, title: `${title} (${itemCount})`, disabled: !itemCount};
                            }).map(({title, dataKey, route, filter, disabled}, index, items) => {
                                const onCounterClick = () => {
                                    setRuntimeScanFilter(filtersDispatch, {items, current: dataKey, minimalSeverity: severity, severityKey: tablesFilter});
                                    setFilters(filtersDispatch, {type: filter, filters: [{scope: tablesFilter, operator: OPERATORS.gte.value, value: [severity]}], isSystem: false});
                                    navigate(route);
                                }
            
                                return (
                                    <Button key={dataKey} disabled={disabled} onClick={onCounterClick}>{title}</Button>
                                )
                            })
                        }
                        {loading && <Loader small absolute={false} />}
                    </div>
            </StepDisplay>
        </React.Fragment>
    )
}

export default SeverityFilterAndCountersSteps;