import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { TagsList } from 'components/Tag';
import { ContainerImageInfoDetails } from './ContainerImageInfoDetails';
import { FieldSet } from 'components/FieldSet';
import { formatTagsToStringsList } from 'utils/utils';


export const ContainerInfoDetails = ({assetData}) => {
    const {containerName, containerID, image, labels} = assetData.assetInfo || {};

    return (
        <>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Container Name">{containerName}</TitleValueDisplay>
                <TitleValueDisplay title="Container ID">{containerID}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Labels"><TagsList items={formatTagsToStringsList(labels)} /></TitleValueDisplay>
            </TitleValueDisplayRow>

            <FieldSet legend="Image">
                <ContainerImageInfoDetails assetData={{assetInfo: image}} />
            </FieldSet>
        </>
    )
}
