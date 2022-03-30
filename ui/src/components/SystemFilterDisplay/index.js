import React from 'react';
import { useNavigate } from 'react-router-dom';
import { isEmpty } from 'lodash';
import CloseButton from 'components/CloseButton';
import { OPERATORS } from 'components/Filter';
import { useFilterDispatch, setFilters } from 'context/FiltersProvider';

import './system-filter-display.scss';

const SystemFilterDisplay = ({onClose, displayText, runtimeScanData}) => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const {items: runtimeScanItems, current, minimalSeverity} = runtimeScanData || {};
    const linkItems = (runtimeScanItems || []).filter(({dataKey}) => dataKey !== current);

    return (
        <div className="system-filter-wrapper">
            <div className="system-filter-content">
                <div className="filter-content">{displayText}</div>
                <div className="filter-close-wrapper" onClick={onClose}>
                    <CloseButton small />
                    <div>Clear filter</div>
                </div>
            </div>
            <div className="system-filter-links">
                {!isEmpty(linkItems) && <div>More scan results:</div>}
                {
                    linkItems.map(({title, filter, dataKey, route}) => {
                        const onClick = () => {
                            setFilters(filtersDispatch, {type: filter, filters: {currentRuntimeScan: {items: runtimeScanItems, current: dataKey, minimalSeverity}}, isSystem: true});
                            setFilters(filtersDispatch, {type: filter, filters: [{scope: "vulnerabilitySeverity", operator: OPERATORS.gte.value, value: [minimalSeverity]}], isSystem: false});
                            navigate(route);
                        }

                        return (
                            <div key={title} className="system-filter-link" onClick={onClick}>{title}</div>
                        )
                    })
                }
            </div>
        </div>
    )
}

export default SystemFilterDisplay;



