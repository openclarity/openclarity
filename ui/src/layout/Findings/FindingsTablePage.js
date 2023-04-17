import React from 'react';
import TablePage from 'components/TablePage';
import { getAssetColumnsFiltersConfig, scanColumnsFiltersConfig } from 'utils/utils';
import { APIS } from 'utils/systemConsts';
import { FILTER_TYPES } from 'context/FiltersProvider';

const FindingsTablePage = ({tableTitle, findingsObjectType, columns, filterType, filtersConfig}) => (
    <TablePage
        columns={columns}
        url={APIS.FINDINGS}
        tableTitle={tableTitle}
        filterType={filterType}
        filtersConfig={[
            ...filtersConfig,
            ...getAssetColumnsFiltersConfig({prefix: "asset.targetInfo", withType: false}),
            ...scanColumnsFiltersConfig
        ]}
        systemFilterType={FILTER_TYPES.FINDINGS_GENERAL}
        filters={`findingInfo.objectType eq '${findingsObjectType}'`}
        expand="asset,scan"
        defaultSortBy={{sortIds: ["scan.startTime"], desc: true}}
    />
)

export default FindingsTablePage;