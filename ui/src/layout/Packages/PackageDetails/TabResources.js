import React, { useMemo } from 'react';
import Table from 'components/Table';
import InnerAppLink from 'components/InnerAppLink';
import Filter, { OPERATORS, formatFiltersToQueryParams } from 'components/Filter';
import VerticalItemsList from 'components/VerticalItemsList';
import { useFilterState, useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { ROUTES } from 'utils/systemConsts';

const TABLE_TITLE = "Application Resources";

const TabResources = ({id, refreshTimestamp}) => {
    const filtersState = useFilterState();
    const {tableFilters: filters} = filtersState[FILTER_TYPES.PACKAGE_RESOURCES];
    const filtersDispatch = useFilterDispatch();

    const columns = useMemo(() => [
        {
            Header: "Resource Name",
            id: "resourceName",
            accessor: "resourceName"
        },
        {
            Header: "Resource Hash",
            id: "resourceHash",
            Cell: ({row}) => {
                const {resourceHash} = row.original;
                
                return (
                    <InnerAppLink pathname={ROUTES.APPLICATION_RESOURCES} onClick={() => {
                        setFilters(filtersDispatch, {
                            type: FILTER_TYPES.APPLICATION_RESOURCES,
                            filters: [{scope: "resourceHash", operator: OPERATORS.is.value, value: [resourceHash]}]
                        });
                    }}>{resourceHash}</InnerAppLink>
                )
            },
            width: 200,
            canSort: true
        },
        {
            Header: "SBOM Analyzers",
            id: "reportingSBOMAnalyzers",
            Cell: ({row}) => <VerticalItemsList items={row.original.reportingSBOMAnalyzers} />
        }
    ], [filtersDispatch]);
    
    return (
        <div className="package-tab-resources">
            <Filter
                filters={filters}
                onFilterUpdate={filters => setFilters(filtersDispatch, {type: FILTER_TYPES.PACKAGE_RESOURCES, filters, isSystem: false})}
                filtersMap={{
                    resourceName: {value: "resourceName", label: "Resource name", operators: [
                        {...OPERATORS.is, valueItems: [], creatable: true},
                        {...OPERATORS.isNot, valueItems: [], creatable: true},
                        {...OPERATORS.start},
                        {...OPERATORS.end},
                        {...OPERATORS.contains, valueItems: [], creatable: true}
                    ]},
                    resourceHash: {value: "resourceHash", label: "Resource hash", operators: [
                        {...OPERATORS.is, valueItems: [], creatable: true},
                        {...OPERATORS.isNot, valueItems: [], creatable: true},
                        {...OPERATORS.start},
                        {...OPERATORS.end},
                        {...OPERATORS.contains, valueItems: [], creatable: true}
                    ]},
                    reportingSBOMAnalyzers: {value: "reportingSBOMAnalyzers", label: "SBOM Analyzers", operators: [
                        {...OPERATORS.containElements, valueItems: [], creatable: true},
                        {...OPERATORS.doesntContainElements, valueItems: [], creatable: true}
                    ]}
                }}
                isSmall
            />
            <Table
                columns={columns}
                paginationItemsName={TABLE_TITLE.toLowerCase()}
                url={`packages/${id}/applicationResources`}
                filters={formatFiltersToQueryParams(filters)}
                noResultsTitle={TABLE_TITLE}
                defaultSortBy={[{id: "resourceName", desc: true}]}
                refreshTimestamp={refreshTimestamp}
            />
        </div>
    )
}

export default TabResources;