import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';


export const DirInfoDetails = ({assetData}) => {
    const {dirName, location} = assetData.assetInfo || {};

    return (
        <TitleValueDisplayRow>
            <TitleValueDisplay title="Dir Name">{dirName}</TitleValueDisplay>
            <TitleValueDisplay title="Location">{location}</TitleValueDisplay>
        </TitleValueDisplayRow>
    )
}
