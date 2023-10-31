import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { formatDate } from 'utils/utils';

export const CommonAssetMetadata = ({assetData: {firstSeen, lastSeen, terminatedOn, assetInfo}}) =>
    <>
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Type">{assetInfo.objectType}</TitleValueDisplay>
            <TitleValueDisplay title="First Seen">{formatDate(firstSeen)}</TitleValueDisplay>
        </TitleValueDisplayRow>
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Last Seen">{formatDate(lastSeen)}</TitleValueDisplay>
            <TitleValueDisplay title="Terminated On">{formatDate(terminatedOn)}</TitleValueDisplay>
        </TitleValueDisplayRow>
    </>
