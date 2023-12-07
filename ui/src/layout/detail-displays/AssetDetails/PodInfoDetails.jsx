import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';


export const PodInfoDetails = ({assetData}) => {
    const {podName, location} = assetData.assetInfo || {};

    return (
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Pod Name">{podName}</TitleValueDisplay>
            <TitleValueDisplay title="Location">{location}</TitleValueDisplay>
        </TitleValueDisplayRow>
    )
}
