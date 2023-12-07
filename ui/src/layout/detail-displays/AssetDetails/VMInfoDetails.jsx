import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { TagsList } from 'components/Tag';
import { formatDate, formatTagsToStringsList } from 'utils/utils';

export const VMInfoDetails = ({assetData}) => {
    const {instanceID, location, tags, image, instanceType, platform, launchTime, rootVolume} = assetData.assetInfo || {};
    const {sizeGB, encrypted} = rootVolume || {};

    return (
        <>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Instance ID">{instanceID}</TitleValueDisplay>
                <TitleValueDisplay title="Location">{location}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Tags"><TagsList items={formatTagsToStringsList(tags)} /></TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Image">{image}</TitleValueDisplay>
                <TitleValueDisplay title="Instance type">{instanceType}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Platform">{platform}</TitleValueDisplay>
                <TitleValueDisplay title="Launch time">{formatDate(launchTime)}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Root Volume Size">{sizeGB} GB</TitleValueDisplay>
                <TitleValueDisplay title="Encrypted Root Volume">{encrypted}</TitleValueDisplay>
            </TitleValueDisplayRow>
        </>
    )
}
