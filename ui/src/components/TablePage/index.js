import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { isEmpty } from 'lodash';
import ContentContainer from 'components/ContentContainer';
import Table from 'components/Table';
import SystemFilterBanner from 'components/SystemFiltersBanner';
import Filter, { formatFiltersToOdataItems } from 'components/Filter';
import { toCapitalized, BoldText } from 'utils/utils';
import { useFilterState, useFilterDispatch, resetSystemFilters, setPage, setSort, setFilters } from 'context/FiltersProvider';

const TablePage = (props) => {
    const {tableTitle, filterType, systemFilterType, filters, expand, select, withMargin, defaultSortBy: initialSortBy,
        filtersConfig, customHeaderDisplay: CustomHeaderDisplay, ...tableProps} = props;
    
    const navigate = useNavigate();
    const {pathname} = useLocation();

    const filtersDispatch = useFilterDispatch();
    const filtersState = useFilterState();

    const {systemFilters} = filtersState[systemFilterType || filterType];
    const {name: systemFilterName, suffix: systemSuffix, backPath: systemFilterBackPath, filter: systemFilter, customDisplay} = systemFilters;

    const {tableFilters, selectedPageIndex, tableSort} = filtersState[filterType];

    const setTableFilters = (filters) => setFilters(filtersDispatch, {type: filterType, filters, isSystem: false});
    
    const onSystemFilterClose = () => resetSystemFilters(filtersDispatch, systemFilterType || filterType);
    
    const fitlersList = [
        ...(!!filters ? [filters] : []),
        ...(!!tableFilters ? formatFiltersToOdataItems(tableFilters)  : []),
        ...(!!systemFilter ? [systemFilter]  : [])
    ];
    
    return (
        <div style={!!withMargin && !!systemFilterName ? {marginTop: "80px"} : {}}>
            {!!systemFilterName &&
                <SystemFilterBanner
                    displayText={<span>{`${toCapitalized(tableTitle)} for `}<BoldText>{systemFilterName}</BoldText>{` ${systemSuffix}`}</span>}
                    onClose={onSystemFilterClose}
                    backPath={systemFilterBackPath}
                    customDisplay={customDisplay}
                />
            }
            <div style={{...(!!withMargin ? {margin: "30px 30px 35px 30px"} : {marginBottom: "35px"}), position: "relative", visibility: (!!filtersConfig || !!CustomHeaderDisplay) ? "visible" : "hidden"}}>
                {!!filtersConfig &&
                    <Filter
                        filters={tableFilters}
                        onFilterUpdate={filters => setTableFilters(filters)}
                        filtersConfig={filtersConfig}
                    />
                }
                {!!CustomHeaderDisplay && <div style={{position: "absolute", top: 0, left: "100px"}}><CustomHeaderDisplay /></div>}
            </div>
            <ContentContainer withMargin={withMargin}>
                <Table
                    paginationItemsName={tableTitle.toLowerCase()}
                    filters={{
                        ...(!!expand ? {"$expand": expand} : {}),
                        ...(!!select ? {"$select": select} : {}),
                        ...(fitlersList.length > 0 ? {"$filter": fitlersList.join(" and ")} : {})
                    }}
                    noResultsTitle={tableTitle}
                    onLineClick={({id}) => navigate(`${pathname}/${id}`)}
                    defaultPageIndex={selectedPageIndex}
                    onPageChange={pageIndex => setPage(filtersDispatch, {type: filterType, pageIndex})}
                    defaultSortBy={isEmpty(tableSort) ? initialSortBy : tableSort}
                    onSortChnage={tableSort => setSort(filtersDispatch, {type: filterType, tableSort})}
                    showCustomEmptyDisplay={isEmpty(tableFilters)}
                    {...tableProps}
                />
            </ContentContainer>
        </div>
    )
}

export default TablePage;