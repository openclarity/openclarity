import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import { TagsList } from 'components/Tag';
import { formatTagsToStringsList } from 'utils/utils';
import prettyBytes from 'pretty-bytes';


export const ContainerImageInfoDetails = ({assetData}) => {
    const {imageID, repoDigests, repoTags, labels, architecture, os, size} = assetData.assetInfo || {};

    return (
        <>
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Image ID">{imageID}</TitleValueDisplay>
                <TitleValueDisplay title="Size">{prettyBytes(size)}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Repo digests">{(repoDigests ?? []).map(digest => <div>{digest}</div>)}</TitleValueDisplay>
            </TitleValueDisplayRow>

            <TitleValueDisplayRow>
                <TitleValueDisplay title="Repo tags">{(repoTags ?? []).map(tag => <div>{tag}</div>)}</TitleValueDisplay>
            </TitleValueDisplayRow>
            
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Architecture">{architecture}</TitleValueDisplay>
                <TitleValueDisplay title="OS">{os}</TitleValueDisplay>
            </TitleValueDisplayRow>
            
            <TitleValueDisplayRow>
                <TitleValueDisplay title="Labels"><TagsList items={formatTagsToStringsList(labels)} /></TitleValueDisplay>
            </TitleValueDisplayRow>
        </>
    )
}
