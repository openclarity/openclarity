import React, { useState } from 'react';
import classnames from 'classnames';
import Tabs from 'components/Tabs';
import IconWithTooltip from 'components/IconWithTooltip';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import WidgetWrapper from '../WidgetWrapper';

import COLORS from 'utils/scss_variables.module.scss';

const FINDINGS_ITEMS = [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING).filter(({value}) => value !== FINDINGS_MAPPING.PACKAGES.value)];

const Tab = ({widgetName, title, icon, isActive}) => (
    <IconWithTooltip
        tooltipId={`${widgetName}-${title}`}
        tooltipText={title}
        name={icon}
        style={{color: isActive ? COLORS["color-grey-black"] : COLORS["color-grey"]}}
        size={20}
    />
)

const FindingsTabsWidget = ({widgetName, className, title, tabContent: TabContent}) => {
    const WIDGET_TAB_ITEMS = FINDINGS_ITEMS.map(({dataKey, icon, title}) => (
        {id: dataKey, customTitle: ({isActive}) => <Tab widgetName={widgetName} title={title} icon={icon} isActive={isActive} />}
    ))

    const [selectedTabId, setSelectedTabId] = useState(WIDGET_TAB_ITEMS[0].id);

    return (
        <WidgetWrapper title={title} className={classnames("findings-tabs-widget", className)} titleMargin={10}>
            <Tabs
                items={WIDGET_TAB_ITEMS}
                checkIsActive={({id}) => id === selectedTabId}
                onClick={({id}) => setSelectedTabId(id)}
                tabItemPadding={15}
            />
            <div>{!!TabContent ? <TabContent selectedTabId={selectedTabId} /> : <div style={{margin: "10px 0"}}>TBD</div>}</div>
        </WidgetWrapper>
    )
}

export default FindingsTabsWidget;
