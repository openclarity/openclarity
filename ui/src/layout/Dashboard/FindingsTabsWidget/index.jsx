import React, { useState } from 'react';
import classnames from 'classnames';
import { get } from 'lodash';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import Tabs from 'components/Tabs';
import IconWithTooltip from 'components/IconWithTooltip';
import WidgetWrapper from '../WidgetWrapper';

import COLORS from 'utils/scss_variables.module.scss';

import './findings-tabs-widget.scss';

const WidgetContent = ({data=[], getHeaderItems, getBodyItems, selectedId}) => {
    const displayData = (data || []).slice(0, 5);

    return (
        <table className="tabbed-widget-table">
            <thead>
                <tr>
                    {getHeaderItems(selectedId).map((item, index, items) => (
                        <th key={index} style={items.length - 1 === index ? {textAlign: "right"} : {}}>{item}</th>
                    ))}
                </tr>
            </thead>
            <tbody>
                {
                    displayData.length > 0 ?
                        displayData.map((item, index) => {
                            return (
                                <tr key={index}>
                                    {getBodyItems(selectedId).map(({dataKey, customDisplay: CustomDisplay}, index, items) => (
                                        <td key={index} style={items.length - 1 === index ? {textAlign: "right"} : {}}>
                                            {!!CustomDisplay ? <CustomDisplay {...item} /> : get(item, dataKey)}
                                        </td>
                                    ))}
                                </tr>
                            )
                        })
                        : <td><tr><div className="empty-results-display-wrapper">No results available</div></tr></td>
                }
            </tbody>
        </table>
    )
}

const Tab = ({widgetName, title, icon, isActive}) => (
    <IconWithTooltip
        tooltipId={`${widgetName}-${title}`}
        tooltipText={title}
        name={icon}
        style={{color: isActive ? COLORS["color-grey-black"] : COLORS["color-grey"]}}
        size={20}
    />
)

const FindingsTabsWidget = ({widgetName, findingsItems, className, title, url, getHeaderItems, getBodyItems}) => {
    const [{data, error, loading}] = useFetch(url, {urlPrefix: "ui"});
    
    const WIDGET_TAB_ITEMS = findingsItems.map(({dataKey, icon, title}) => (
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
            <div className="tabbed-widget-table-wrapper">
                {
                    loading ? <Loader /> : (error ? null :
                        <WidgetContent
                            data={!!data ? data[selectedTabId] : []}
                            getHeaderItems={getHeaderItems}
                            getBodyItems={getBodyItems}
                            selectedId={selectedTabId}
                        />
                    )
                }
            </div>
        </WidgetWrapper>
    )
}

export default FindingsTabsWidget;
