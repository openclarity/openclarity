import moment from "moment";
import cronstrue from "cronstrue";
import CVSS from "@turingpointde/cvss.js";
import { isEmpty, orderBy } from "lodash";
import {
  FINDINGS_MAPPING,
  VULNERABILITY_FINDINGS_ITEM,
} from "utils/systemConsts";
import IconWithTooltip from "components/IconWithTooltip";
import VulnerabilitiesDisplay, {
  VULNERABILITY_SEVERITY_ITEMS,
} from "components/VulnerabilitiesDisplay";
import { OPERATORS } from "components/Filter";

export const formatDateBy = (date, format) =>
  !!date ? moment(date).format(format) : "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const calculateDuration = (startTime, endTime) => {
  const startMoment = moment(startTime);
  const endMoment = moment(endTime || new Date());

  if (!startTime) {
    return null;
  }

  const range = ["days", "hours", "minutes", "seconds"]
    .map((item) => ({ diff: endMoment.diff(startMoment, item), label: item }))
    .find(({ diff }) => diff > 1);

  return !!range ? `${range.diff} ${range.label}` : "less than 1 second";
};

export const toCapitalized = (string) =>
  string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();

export const BoldText = ({ children, style = {} }) => (
  <span style={{ fontWeight: "bold", ...style }}>{children}</span>
);

export const cronExpressionToHuman = (value) =>
  cronstrue.toString(value, { use24HourTimeFormat: true });

export const formatNumber = (value) =>
  new Intl.NumberFormat("en-US").format(parseInt(value || 0, 10));

export const getHighestVersionCvssData = (cvssData) => {
  if (isEmpty(cvssData)) {
    return {};
  }

  const sortedCvss = orderBy(cvssData || [], ["version"], ["desc"]);

  const { vector, metrics, version } = sortedCvss[0];

  const serverData = {
    vector,
    score: metrics.baseScore,
    exploitabilityScore: metrics.exploitabilityScore,
    impactScore: metrics.impactScore,
  };

  if (version === "2.0") {
    return serverData;
  }

  const cvssVector = CVSS(vector);

  return {
    ...serverData,
    temporalScore: cvssVector.getTemporalScore(),
    environmentalScore: cvssVector.getEnvironmentalScore(),
    severity: cvssVector.getRating(),
    metrics: cvssVector.getDetailedVectorObject().metrics,
  };
};

export const getFindingsColumnsConfigList = (props) =>
  Object.values(FINDINGS_MAPPING).map(({ totalKey, title, icon }) => {
    const { tableTitle, withAssetPrefix = false } = props;
    const prefix = withAssetPrefix ? "asset." : "";

    return {
      Header: (
        <IconWithTooltip
          tooltipId={`table-header-${tableTitle}-${totalKey}`}
          tooltipText={title}
          name={icon}
        />
      ),
      id: totalKey,
      sortIds: [`${prefix}summary.${totalKey}`],
      accessor: (original) => {
        const { summary } = withAssetPrefix ? original.asset : original;

        return isEmpty(summary) ? 0 : formatNumber(summary[totalKey] || 0);
      },
      width: 50,
    };
  });

export const getVulnerabilitiesColumnConfigItem = (props) => {
  const { tableTitle, withAssetPrefix = false } = props;
  const prefix = withAssetPrefix ? "asset." : "";
  const { title: vulnerabilitiesTitle, icon: vulnerabilitiesIcon } =
    VULNERABILITY_FINDINGS_ITEM;

  return {
    Header: (
      <IconWithTooltip
        tooltipId={`table-header-${tableTitle}-vulnerabilities`}
        tooltipText={vulnerabilitiesTitle}
        name={vulnerabilitiesIcon}
      />
    ),
    id: "vulnerabilities",
    sortIds: [
      `${prefix}summary.totalVulnerabilities.totalCriticalVulnerabilities`,
      `${prefix}summary.totalVulnerabilities.totalHighVulnerabilities`,
      `${prefix}summary.totalVulnerabilities.totalMediumVulnerabilities`,
      `${prefix}summary.totalVulnerabilities.totalLowVulnerabilities`,
      `${prefix}summary.totalVulnerabilities.totalNegligibleVulnerabilities`,
    ],
    Cell: ({ row }) => {
      const { id, summary } = withAssetPrefix
        ? row.original.asset
        : row.original;

      return (
        <VulnerabilitiesDisplay
          minimizedTooltipId={id}
          counters={summary?.totalVulnerabilities}
          isMinimized
        />
      );
    },
    width: 50,
  };
};

export const findingsColumnsFiltersConfig = Object.values(FINDINGS_MAPPING).map(
  ({ totalKey, title }) => {
    const filterKey = `summary.${totalKey}`;

    return {
      value: filterKey,
      label: title,
      isNumber: true,
      operators: [
        { ...OPERATORS.eq, valueItems: [], creatable: true },
        { ...OPERATORS.ne, valueItems: [], creatable: true },
        { ...OPERATORS.ge },
        { ...OPERATORS.le },
      ],
    };
  },
);

export const vulnerabilitiesCountersColumnsFiltersConfig = Object.values(
  VULNERABILITY_SEVERITY_ITEMS,
).map(({ totalKey, title }) => {
  const filterKey = `summary.totalVulnerabilities.${totalKey}`;

  return {
    value: filterKey,
    label: `${title} vulnerabilities`,
    isNumber: true,
    operators: [
      { ...OPERATORS.eq, valueItems: [], creatable: true },
      { ...OPERATORS.ne, valueItems: [], creatable: true },
      { ...OPERATORS.ge },
      { ...OPERATORS.le },
    ],
  };
});

export const scanColumnsFiltersConfig = [
  {
    value: "scan.name",
    label: "Scan name",
    operators: [
      { ...OPERATORS.eq, valueItems: [], creatable: true },
      { ...OPERATORS.ne, valueItems: [], creatable: true },
      { ...OPERATORS.startswith },
      { ...OPERATORS.endswith },
      { ...OPERATORS.contains, valueItems: [], creatable: true },
    ],
  },
  {
    value: "scan.endTime",
    label: "Scan end time",
    isDate: true,
    operators: [{ ...OPERATORS.ge }, { ...OPERATORS.le }],
  },
];

export const getAssetColumnsFiltersConfig = (props) => {
  const {
    prefix = "assetInfo",
    withType = true,
    withLabels = true,
  } = props || {};

  const ASSET_TYPE_ITEMS = [
    { value: "VMInfo", label: "VMInfo" },
    { value: "ContainerInfo", label: "ContainerInfo" },
    { value: "ContainerImageInfo", label: "ContainerImageInfo" },
    { value: "PodInfo", label: "PodInfo" },
    { value: "DirInfo", label: "DirInfo" },
  ];

  return [
    {
      value: `${prefix}.instanceID`,
      label: "Asset name",
      operators: [
        { ...OPERATORS.eq, valueItems: [], creatable: true },
        { ...OPERATORS.ne, valueItems: [], creatable: true },
        { ...OPERATORS.startswith },
        { ...OPERATORS.endswith },
        { ...OPERATORS.contains, valueItems: [], creatable: true },
      ],
    },
    ...(!withLabels
      ? []
      : [
          {
            value: `${prefix}.tags`,
            label: "Labels",
            operators: [
              { ...OPERATORS.contains, valueItems: [], creatable: true },
            ],
          },
        ]),
    ...(!withType
      ? []
      : [
          {
            value: `${prefix}.objectType`,
            label: "Asset type",
            operators: [
              { ...OPERATORS.eq, valueItems: ASSET_TYPE_ITEMS },
              { ...OPERATORS.ne, valueItems: ASSET_TYPE_ITEMS },
            ],
          },
        ]),
    {
      value: `${prefix}.location`,
      label: "Asset location",
      operators: [
        { ...OPERATORS.eq, valueItems: [], creatable: true },
        { ...OPERATORS.ne, valueItems: [], creatable: true },
        { ...OPERATORS.startswith },
        { ...OPERATORS.endswith },
        { ...OPERATORS.contains, valueItems: [], creatable: true },
      ],
    },
  ];
};

export const formatTagsToStringsList = (tags) =>
  tags?.map(({ key, value }) => `${key}=${value}`);

export function getAssetName(assetInfo) {
  switch (assetInfo.objectType) {
    case "VMInfo":
      return assetInfo.instanceID;
    case "PodInfo":
      return assetInfo.podName;
    case "DirInfo":
      return assetInfo.dirName;
    case "ContainerImageInfo":
      return assetInfo.imageID;
    case "ContainerInfo":
      return assetInfo.containerName;
    default:
      return assetInfo.id;
  }
}
