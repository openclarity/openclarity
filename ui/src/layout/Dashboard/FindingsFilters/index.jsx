import React from 'react';
import IconWithTooltip from 'components/IconWithTooltip';

import COLORS from 'utils/scss_variables.module.scss';

import './findings-filters.scss';

const FilterButton = ({widgetName, icon, title, color, selected, onClick}) => (
    <div className="findings-filter-button" onClick={onClick}>
        <IconWithTooltip
            tooltipId={`findings-filter-${widgetName}-${title}`}
            tooltipText={title}
            name={icon}
            style={{color: selected ? color : COLORS["color-grey"]}}
            size={20}
        />
    </div>
)

const FindingsFilters = ({widgetName, findingsItems, findingKeyName, selectedFilters, setSelectedFilters}) => (
    <div className="findings-filters">
        {
            findingsItems.map((item) => {
                const {color, icon, title} = item;
                const findingKey = item[findingKeyName];

                return ((
                    <FilterButton
                        key={title}
                        widgetName={widgetName}
                        icon={icon}
                        title={title}
                        color={color}
                        selected={selectedFilters.includes(findingKey)}
                        onClick={() => {
                            setSelectedFilters(selectedFilters => {
                                if (selectedFilters.includes(findingKey)) {
                                    return selectedFilters.filter(selectedType => selectedType !== findingKey);
                                } else {
                                    return [...selectedFilters, findingKey];
                                }
                            })
                        }}
                    />
                ))
            })
        }
    </div>
)

export default FindingsFilters;