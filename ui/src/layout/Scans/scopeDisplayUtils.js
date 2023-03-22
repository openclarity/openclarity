import React from 'react';
import { ValuesListDisplay } from 'components/TitleValueDisplay';
import ExpandableList from 'components/ExpandableList';
import { formatRegionsToStrings } from 'layout/Scans/utils';

const SCOPE_ALL = "All";

export const ExpandableScopeDisplay = ({all, regions}) => (
    all ? SCOPE_ALL : <ExpandableList items={formatRegionsToStrings(regions)} />
)

export const ScopeDisplay = ({all, regions}) => {
    if (all) {
        return SCOPE_ALL;
    }

    return ( 
        <ValuesListDisplay values={formatRegionsToStrings(regions)} />
    )
}