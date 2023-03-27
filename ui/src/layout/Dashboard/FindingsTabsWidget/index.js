import React, { useState } from 'react';
import classnames from 'classnames';
import Tabs from 'components/Tabs';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import WidgetWrapper from '../WidgetWrapper';
import { LegendIcon } from '../utils';

import COLORS from 'utils/scss_variables.module.scss';

const Tab = ({widgetName, title, icon, isActive}) => (
    <LegendIcon widgetName={widgetName} title={title} icon={icon} color={isActive ? COLORS["color-grey-black"] : COLORS["color-grey"]} size={20} />
)

const FindingsTabsWidget = ({widgetName, className, title, tabContent: TabContent}) => {
    const {dataKey: vulnerabilitiesKey, title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon} = VULNERABIITY_FINDINGS_ITEM;

    const WIDGET_TAB_ITEMS = [
        {id: vulnerabilitiesKey, customTitle: ({isActive}) => <Tab widgetName={widgetName} title={vulnerabilitiesTitle} icon={vulnerabilitiesIcon} isActive={isActive} />},
        ...Object.keys(FINDINGS_MAPPING).filter(findingType => findingType !== FINDINGS_MAPPING.PACKAGES.value).map(findingType => {
            const {dataKey, icon, title} = FINDINGS_MAPPING[findingType];
    
            return {id: dataKey, customTitle: ({isActive}) => <Tab widgetName={widgetName} title={title} icon={icon} isActive={isActive} />}
        })
    ];

    const [selectedTabId, setSelectedTabId] = useState(WIDGET_TAB_ITEMS[0].id);

    return (
        <WidgetWrapper title={title} className={classnames("findings-tabs-widget", className)}>
            <Tabs
                items={WIDGET_TAB_ITEMS}
                checkIsActive={({id}) => id === selectedTabId}
                onClick={({id}) => setSelectedTabId(id)}
                tabItemPadding={15}
            />
            <div style={{marginTop: "10px"}}>{!!TabContent ? <TabContent selectedTabId={selectedTabId} /> : "TBD"}</div>
        </WidgetWrapper>
    )
}

export default FindingsTabsWidget;
