import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useFetch } from 'hooks';
import { components } from 'react-select';
import { useFilterDispatch, setFilters, FILTERR_TYPES } from 'context/FiltersProvider';
import Icon, { ICON_NAMES } from 'components/Icon';
import { OPERATORS } from 'components/Filter';
import Button from 'components/Button';
import DropdownSelect from 'components/DropdownSelect';
import { ROUTES, SEVERITY_ITEMS } from 'utils/systemConsts';
import StepDisplay from '../StepDisplay';

import './severity-counters-step.scss';

const LINKS_DISPLAY_MAP = [
    {title: "Applications", dataKey: "applications", route: ROUTES.APPLICATIONS, filter: FILTERR_TYPES.APPLICATIONS},
    {title: "Application Resources", dataKey: "resources", route: ROUTES.APPLICATION_RESOURCES, filter: FILTERR_TYPES.APPLICATION_RESOURCES},
    {title: "Packages", dataKey: "packages", route: ROUTES.PACKAGES, filter: FILTERR_TYPES.PACKAGES},
    {title: "Vulnerabilities", dataKey: "vulnerabilities", route: ROUTES.VULNERABILITIES, filter: FILTERR_TYPES.VULNERABILITIES}
];

const SelectItem = ({label, color}) => (
    <div style={{display: "flex", alignItems: "center"}}>
        <Icon name={ICON_NAMES.BUG} style={{color}} />
        <div style={{marginLeft: "5px"}}>{label}</div>
    </div>
)

const SeverityCountersStep = () => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const [{loading, data}, fetchData] = useFetch("runtime/scan/results", {loadOnMount: false});

    const [selectedSeverity, setSelectedSeverity] = useState(SEVERITY_ITEMS.NEGLIGIBLE);

    useEffect(() => {
        fetchData({queryParams: {"vulnerabilitySeverity[gte]": selectedSeverity.value}});
    }, [selectedSeverity, fetchData]);

    const {counters={}} = data || {};

    return (
        <StepDisplay step="3"  title="Minimal severity to view:" className="severity-counters-step">
            <DropdownSelect
                name="runtime-scan-select-severity"
                className="severity-select"
                value={selectedSeverity}
                items={Object.values(SEVERITY_ITEMS)}
                onChange={setSelectedSeverity}
                components={{
                    Option: props => {
                        const {label, color} = props.data;

                        return (
                            <components.Option {...props}>
                                <SelectItem label={label} color={color} />
                            </components.Option>
                        )
                    },
                    SingleValue: props => {
                        const {label, color} = props.data;

                        return (
                            <components.SingleValue {...props}>
                                <SelectItem label={label} color={color} />
                            </components.SingleValue>
                        )
                    }
                }}
            />
            {loading ? null :
                <div className="select-counter-buttons-wrapper">
                    {
                        LINKS_DISPLAY_MAP.map((linkItem) => {
                            const {title, dataKey} = linkItem;
                            const itemCount = counters[dataKey] || 0;

                            return {...linkItem, title: `${title} (${itemCount})`, disabled: !itemCount};
                        }).map(({title, dataKey, route, filter, disabled}, index, items) => {
                            const onCounterClick = () => {
                                setFilters(filtersDispatch, {type: filter, filters: {currentRuntimeScan: {items, current: dataKey, minimalSeverity: selectedSeverity.value}}, isSystem: true});
                                setFilters(filtersDispatch, {type: filter, filters: [{scope: "vulnerabilitySeverity", operator: OPERATORS.gte.value, value: [selectedSeverity.value]}], isSystem: false});
                                navigate(route);
                            }
        
                            return (
                                <Button key={dataKey} disabled={disabled} onClick={onCounterClick}>{title}</Button>
                            )
                        })
                    }
                </div>
            }
        </StepDisplay>
    )
}

export default SeverityCountersStep;