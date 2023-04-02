import React from 'react';
import VulnerabilitiesDisplay, { VULNERABILITY_SEVERITY_ITEMS } from 'components/VulnerabilitiesDisplay';
import { APIS, VULNERABIITY_FINDINGS_ITEM, FINDINGS_MAPPING } from 'utils/systemConsts';
import { formatNumber } from 'utils/utils';
import FindingsTabsWidget from '../FindingsTabsWidget';

const FINDINGS_ITEMS = [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING).filter(({value}) => value !== FINDINGS_MAPPING.PACKAGES.value)];

const RiskiestAssetsWidget = ({className}) => (
    <FindingsTabsWidget
        className={className}
        findingsItems={FINDINGS_ITEMS}
        title="Riskiest assets"
        widgetName="riskiers-assets"
        url={APIS.DASHBOARD_RISKIEST_ASSETS}
        getHeaderItems={() => (["Name", "Type", "Findings"])}
        getBodyItems={(selectedId) => ([
            {dataKey: "assetInfo.name"},
            {dataKey: "assetInfo.type"},
            {customDisplay: ({count, assetInfo, ...props}) => {
                if (selectedId === VULNERABIITY_FINDINGS_ITEM.dataKey) {
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
                
                return formatNumber(count);
            }}
        ])}
    />
)

export default RiskiestAssetsWidget;