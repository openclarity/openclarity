import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';

const LinksDisplay = ({title, links}) => (
    <TitleValueDisplayRow>
        <TitleValueDisplay title={title} withOpen defaultOpen>
            {
                links?.map((link, index) => (
                    <div key={index}><a href={link} target="_blank" rel="noopener noreferrer">{link}</a></div>
                ))
            }
        </TitleValueDisplay>
    </TitleValueDisplayRow>
)

export default LinksDisplay;
