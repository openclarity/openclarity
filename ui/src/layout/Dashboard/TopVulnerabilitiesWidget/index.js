import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { isEmpty, isUndefined } from 'lodash';
import classnames from 'classnames';
import VulnerabilitiesSummaryDisplay from 'components/VulnerabilitiesSummaryDisplay';
import { TooltipWrapper } from 'components/Tooltip';
import { OPERATORS } from 'components/Filter';
import { FILTER_TYPES } from 'context/FiltersProvider';
import { useFilterDispatch, setFilters } from 'context/FiltersProvider';
import { setVulnerabilitiesSystemFilters } from 'layout/Vulnerabilities';
import { ROUTES } from 'utils/systemConsts';
import { BoldText } from 'utils/utils';
import WidgetWrapper from '../WidgetWrapper';
import { NO_DATA } from '../utils';

import './top-vulnerabilities-widget.scss';

const WIDGET_TABS = {
    APPLICATIONS: {value: "applications", nameProp: "applicationName", systemFilter: "applicationID", filterTitle: "application"},
    RESOURCES: {value: "resources", nameProp: "resourceName", systemFilter: "applicationResourceID", filterTitle: "resource"},
    PACKAGES: {
        value: "packages",
        getName: ({packageName, version}) => isUndefined(packageName) ? NO_DATA : `${packageName} ${version}`,
        getCustomFilterData: ({packageName, version}) => [
            {scope: "packageVersion", operator: OPERATORS.is.value, value: [version]},
            {scope: "packageName", operator: OPERATORS.is.value, value: [packageName]}
        ]
    },
}

const NO_DATA_ITEM = {
    [WIDGET_TABS.APPLICATIONS.nameProp]: NO_DATA,
    [WIDGET_TABS.RESOURCES.nameProp]: NO_DATA,
    vulnerabilities: []
}

const TopVulnerabilitiesWidget = ({data}) => {
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const [selectedTab, setSelectedTab] = useState(WIDGET_TABS.APPLICATIONS);

    const noData = isEmpty(data);
    
    return (
        <div>
            <div className="widget-tabs-header">
                {Object.values(WIDGET_TABS).map(tab => {
                    const {value} = tab;

                    return (
                        <div key={value} className={classnames("widget-tab-item", {selected: value === selectedTab.value})} onClick={() => setSelectedTab(tab)}>
                            {value}
                        </div>
                    )
                })}
            </div>
            <div className="widget-tab-content">
                {
                    ((data || {})[selectedTab.value] || [NO_DATA_ITEM]).map((item, index) => {
                        const {nameProp, getName, systemFilter, filterTitle, getCustomFilterData} = selectedTab;
                        const title = !!getName ? getName(item) : item[nameProp];
                        
                        const onItemClick = () => {
                            if (!!getCustomFilterData) {
                                setFilters(filtersDispatch, {type: FILTER_TYPES.VULNERABILITIES, filters: getCustomFilterData(item), isSystem: false});
                            } else {
                                setVulnerabilitiesSystemFilters(filtersDispatch, {[systemFilter]: item.id, title: <span>{`${filterTitle}: `}<BoldText>{title}</BoldText></span>});
                            }
                            
                            navigate(ROUTES.VULNERABILITIES);
                        }

                        return (
                            <div key={index} className={classnames("vulneability-item", {clickable: !noData})} onClick={noData ? undefined : onItemClick}>
                                <TooltipWrapper className="vulneability-item-title" tooltipId={`dashboard-top-line-${index}`} tooltipText={title}>
                                    {title}
                                </TooltipWrapper>
                                <VulnerabilitiesSummaryDisplay id={`dashboard-top-vul-widget-${index}`} vulnerabilities={item.vulnerabilities} withTotal isNarrow />
                            </div>
                        )
                    })
                }
            </div>
        </div>
    )
}

const TopVulnerabilitiesWidgetWrapper = ({refreshTimestamp}) => (
    <WidgetWrapper
        className="top-vulneraboilities-widget"
        title="Top 5 vulnerable elements"
        url="dashboard/mostVulnerable"
        widget={TopVulnerabilitiesWidget}
        refreshTimestamp={refreshTimestamp}
    />
)

export default TopVulnerabilitiesWidgetWrapper;