import React from "react";
import TitleValueDisplay, {
  TitleValueDisplayRow,
} from "components/TitleValueDisplay";
import DoublePaneDisplay from "components/DoublePaneDisplay";
import { FindingsDetailsCommonFields } from "../utils";
import { MISCONFIGURATION_SEVERITY_MAP } from "./utils";
import AssetCountDisplay from "../AssetCountDisplay";

const TabMisconfigurationDetails = ({ data }) => {
  const { id: findingId, findingInfo, firstSeen, lastSeen } = data;
  const {
    id,
    severity,
    description,
    location,
    remediation,
    category,
    message,
  } = findingInfo;

  return (
    <DoublePaneDisplay
      leftPaneDisplay={() => (
        <>
          <TitleValueDisplayRow>
            <TitleValueDisplay title="ID">{id}</TitleValueDisplay>
            <TitleValueDisplay title="Severity">
              {MISCONFIGURATION_SEVERITY_MAP[severity]}
            </TitleValueDisplay>
          </TitleValueDisplayRow>
          <TitleValueDisplayRow>
            <TitleValueDisplay title="File path">{location}</TitleValueDisplay>
          </TitleValueDisplayRow>
          <TitleValueDisplayRow>
            <TitleValueDisplay title="Category">{category}</TitleValueDisplay>
            <TitleValueDisplay title="Remediation">
              {remediation}
            </TitleValueDisplay>
          </TitleValueDisplayRow>
          <TitleValueDisplayRow>
            <TitleValueDisplay title="Message" withOpen defaultOpen>
              {message}
            </TitleValueDisplay>
          </TitleValueDisplayRow>
          <TitleValueDisplayRow>
            <TitleValueDisplay title="Description" withOpen defaultOpen>
              {description}
            </TitleValueDisplay>
          </TitleValueDisplayRow>
          <FindingsDetailsCommonFields
            firstSeen={firstSeen}
            lastSeen={lastSeen}
          />
          {AssetCountDisplay(findingId)}
        </>
      )}
    />
  );
};

export default TabMisconfigurationDetails;
