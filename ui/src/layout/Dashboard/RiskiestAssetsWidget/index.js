import React from 'react';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import VulnerabilitiesDisplay, { VULNERABILITY_SEVERITY_ITEMS } from 'components/VulnerabilitiesDisplay';
import { APIS, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import FindingsTabsWidget from '../FindingsTabsWidget';

import './riskiest-assets-widget.scss';

const WidgetContent = ({data=[], maxItems, getCountDisplay}) => {
    const displayData = data.slice(0, maxItems);

    return (
        <table className="tabbed-widget-table">
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Findings</th>
                </tr>
            </thead>
            <tbody>
                {
                    displayData.map((item, index) => {
                        const {name, type} = item?.assetInfo || {};
                        
                        return (
                            <tr key={index}>
                                <td>{name}</td>
                                <td>{type}</td>
                                <td style={{textAlign: "right"}}>{getCountDisplay(item || {})}</td>
                            </tr>
                        )
                    })
                }
            </tbody>
        </table>
    )
}

const RiskiestAssetsWidget = ({className, maxItems=5}) => {
    const [{data, error, loading}] = useFetch(APIS.DASHBOARD_RISKIEST_ASSETS, {urlPrefix: "ui"});

    return (
        <FindingsTabsWidget
            className={className}
            title="Riskiest assets"
            widgetName="riskiers-assets"
            tabContent={({selectedTabId}) => (
                loading ? <Loader absolute={false} /> : (error ? null :
                    <WidgetContent
                        data={!!data ? data[selectedTabId] : []}
                        maxItems={maxItems}
                        getCountDisplay={({count, assetInfo, ...props}) => {
                            if (selectedTabId === VULNERABIITY_FINDINGS_ITEM.dataKey) {
                                const counters = Object.values(VULNERABILITY_SEVERITY_ITEMS).reduce((acc, curr) => {
                                    const {totalKey, countKey} = curr;

                                    return {...acc, [totalKey]: props[countKey]};
                                }, {});

                                return (
                                    <VulnerabilitiesDisplay
                                        minimizedTooltipId={`riskiest-assets-widget-${assetInfo?.name}`}
                                        counters={counters}
                                        isMinimized
                                    />
                                )
                            }
                            
                            return count;
                        }}
                    />
                )
            )}
        />
    )
}

export default RiskiestAssetsWidget;