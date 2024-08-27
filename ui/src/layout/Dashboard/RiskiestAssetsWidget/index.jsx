import React from "react";
import VulnerabilitiesDisplay, {
  VULNERABILITY_SEVERITY_ITEMS,
} from "components/VulnerabilitiesDisplay";
import {
  VULNERABILITY_FINDINGS_ITEM,
  FINDINGS_MAPPING,
} from "utils/systemConsts";
import { formatNumber } from "utils/utils";
import { useQuery } from "@tanstack/react-query";
import FindingsTabsWidget from "../FindingsTabsWidget";
import QUERY_KEYS from "../../../api/constants";
import { openClarityUIBackend } from "../../../api/openClarityApi";

const FINDINGS_ITEMS = [
  VULNERABILITY_FINDINGS_ITEM,
  ...Object.values(FINDINGS_MAPPING).filter(
    ({ value }) => value !== FINDINGS_MAPPING.PACKAGES.value,
  ),
];

const RiskiestAssetsWidget = ({ className }) => {
  const { data, isError, isLoading } = useQuery({
    queryKey: [QUERY_KEYS.riskiestAssets],
    queryFn: () => openClarityUIBackend.dashboardRiskiestAssetsGet(),
  });

  return (
    <FindingsTabsWidget
      className={className}
      findingsItems={FINDINGS_ITEMS}
      title="Riskiest assets"
      widgetName="riskiers-assets"
      getHeaderItems={() => ["Name", "Type", "Findings"]}
      getBodyItems={(selectedId) => [
        { dataKey: "assetInfo.name" },
        { dataKey: "assetInfo.type" },
        {
          customDisplay: ({ count, assetInfo, ...props }) => {
            if (selectedId === VULNERABILITY_FINDINGS_ITEM.dataKey) {
              const counters = Object.values(
                VULNERABILITY_SEVERITY_ITEMS,
              ).reduce((acc, curr) => {
                const { totalKey, countKey } = curr;

                return { ...acc, [totalKey]: props[countKey] };
              }, {});

              return (
                <VulnerabilitiesDisplay
                  minimizedTooltipId={`riskiest-assets-widget-${assetInfo?.name}`}
                  counters={counters}
                  isMinimized
                />
              );
            }

            return formatNumber(count);
          },
        },
      ]}
      data={data?.data}
      isError={isError}
      isLoading={isLoading}
    />
  );
};

export default RiskiestAssetsWidget;
