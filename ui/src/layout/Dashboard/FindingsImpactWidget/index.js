import React from 'react';
import SeverityWithCvssDisplay from 'components/SeverityWithCvssDisplay';
import { getHigestVersionCvssData, formatNumber } from 'utils/utils';
import { FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM, APIS } from 'utils/systemConsts';
import FindingsTabsWidget from '../FindingsTabsWidget';

const FINDINGS_ITEMS = [VULNERABIITY_FINDINGS_ITEM, ...Object.values(FINDINGS_MAPPING)];

const TABS_COLUMNS_MAPPING = {
	[VULNERABIITY_FINDINGS_ITEM.dataKey]: {
		headerItems: ["Name", "Severity"],
		bodyItems: [
			{dataKey: "vulnerability.vulnerabilityName"},
			{customDisplay: ({vulnerability}) => {
        const {severity, cvss, vulnerabilityName} = vulnerability || {};
        const {score, severity: cvssSeverity} = getHigestVersionCvssData(cvss);

        return (
          <SeverityWithCvssDisplay
              severity={severity}
              cvssScore={score}
              cvssSeverity={cvssSeverity?.toUpperCase()}
              compareTooltipId={`severity-compare-tooltip-${vulnerabilityName}`}
          />
        )
      }}
		]
	},
	[FINDINGS_MAPPING.EXPLOITS.dataKey]: {
		headerItems: ["Vulnerability name", "URLs"],
		bodyItems: [
			{dataKey: "exploit.cveID"},
			{dataKey: "exploit.urls"}
		]
	},
	[FINDINGS_MAPPING.MISCONFIGURATIONS.dataKey]: {
		headerItems: ["Message"],
		bodyItems: [
			{dataKey: "misconfiguration.message"}
		]
	},
	[FINDINGS_MAPPING.SECRETS.dataKey]: {
		headerItems: ["Fingerprint"],
		bodyItems: [
			{dataKey: "secret.fingerprint"}
		]
	},
	[FINDINGS_MAPPING.MALWARE.dataKey]: {
		headerItems: ["Malware name"],
		bodyItems: [
			{dataKey: "malware.malwareName"}
		]
	},
	[FINDINGS_MAPPING.ROOTKITS.dataKey]: {
		headerItems: ["Rootkit name", "Message"],
		bodyItems: [
			{dataKey: "rootkit.rootkitName"},
			{dataKey: "rootkit.message"}
		]
	},
	[FINDINGS_MAPPING.PACKAGES.dataKey]: {
		headerItems: ["Package name", "Version"],
		bodyItems: [
			{dataKey: "package.name"},
			{dataKey: "package.version"}
		]
	}
}

const FindingsImpactWidget = ({className}) => (
	<FindingsTabsWidget
		className={className}
		findingsItems={FINDINGS_ITEMS}
		title="Findings impact"
		widgetName="findings-impact"
		url={APIS.DASHBOARD_FINDINGS_IMPACT}
		getHeaderItems={(selectedId) => {
			const {headerItems=[]} = TABS_COLUMNS_MAPPING[selectedId] || {};

			return (
				([...headerItems, "Affected assets"])
			)
		}}
		getBodyItems={(selectedId) => {
			const {bodyItems=[]} = TABS_COLUMNS_MAPPING[selectedId] || {};

			return ([...bodyItems, {customDisplay: ({affectedAssetsCount}) => (formatNumber(affectedAssetsCount))}])
		}}
	/>
)

export default FindingsImpactWidget;