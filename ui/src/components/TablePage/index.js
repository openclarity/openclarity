import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { isEmpty } from 'lodash';
import ContentContainer from 'components/ContentContainer';
import Table from 'components/Table';
import SystemFilterBanner from 'components/SystemFiltersBanner';
import { toCapitalized, BoldText } from 'utils/utils';
import { useFilterState, useFilterDispatch, resetSystemFilters, setPage, setSort } from 'context/FiltersProvider';

const TablePage = ({tableTitle, filterType, systemFilterType, filters, expand, select, withMargin, defaultSortBy: initialSortBy, ...tableProps}) => {
    const navigate = useNavigate();
    const {pathname} = useLocation();

    const filtersState = useFilterState();
    const {systemFilters} = filtersState[systemFilterType || filterType];
    const {selectedPageIndex, tableSort} = filtersState[filterType];
    const filtersDispatch = useFilterDispatch();
    
    const {name: systemFilterName, suffix: systemSuffix, backPath: systemFilterBackPath, filter: systemFilter, customDisplay} = systemFilters;

    const onSystemFilterClose = () => resetSystemFilters(filtersDispatch, systemFilterType || filterType);
    
    const fitlersList = [
        ...(!!filters ? [filters] : []),
        ...(!!systemFilter ? [systemFilter]  : [])
    ]
    
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
                    {...tableProps}
                />
            </ContentContainer>
        </div>
    )
}

export default TablePage;