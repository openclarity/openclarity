import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { ContainerImageInfoDetails } from './ContainerImageInfoDetails';
import { FieldSet } from 'components/FieldSet';


export const ContainerInfoDetails = ({assetData}) => {
    const {containerName, containerID, image} = assetData.assetInfo || {};


    return (
        <>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Container Name">{containerName}</TitleValueDisplay>
                <TitleValueDisplay title="Container ID">{containerID}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <FieldSet legend="Image">
                <ContainerImageInfoDetails assetData={{assetInfo: image}} />
            </FieldSet>
        </>
    )
}
