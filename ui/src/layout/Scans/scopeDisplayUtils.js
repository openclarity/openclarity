import React from 'react';
import { ValuesListDisplay } from 'components/TitleValueDisplay';
import { TagsList } from 'components/Tag';
import ExpandableList from 'components/ExpandableList';
import { formatRegionsToStrings, formatTagsToStringInstances, getEnabledScanTypesList } from 'layout/Scans/utils';

export const ExpandableScopeDisplay = ({all, regions}) => (
    all ? "All" : <ExpandableList items={formatRegionsToStrings(regions)} />
)

export const ScopeDisplay = ({all, regions}) => {
    if (all) {
        return "All";
    }

    return ( 
        <ValuesListDisplay values={formatRegionsToStrings(regions)} />
    )
}

export const ScanTypesDisplay = ({scanFamiliesConfig}) => (
    <ValuesListDisplay values={getEnabledScanTypesList(scanFamiliesConfig)} />
)

export const InstancesDisplay = ({tags}) => (
    <TagsList items={formatTagsToStringInstances(tags)} />
)