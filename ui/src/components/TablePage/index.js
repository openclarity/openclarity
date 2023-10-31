import React, { useEffect } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { isEmpty } from 'lodash';
import { ErrorBoundary } from 'react-error-boundary';
import ContentContainer from 'components/ContentContainer';
import Table from 'components/Table';
import SystemFilterBanner from 'components/SystemFiltersBanner';
import Filter, { formatFiltersToOdataItems } from 'components/Filter';
import Loader from 'components/Loader';
import { toCapitalized, BoldText } from 'utils/utils';
import { useFilterState, useFilterDispatch, resetSystemFilters, setPage, setSort, setFilters, initializeFilters } from 'context/FiltersProvider';

const TablePage = (props) => {
    const {tableTitle, filterType, systemFilterType, filters, expand, select, withMargin, defaultSortBy: initialSortBy,
        filtersConfig, customHeaderDisplay: CustomHeaderDisplay, ...tableProps} = props;
    
    const [searchParams, setSearchParams] = useSearchParams();

    const navigate = useNavigate();
    const {pathname} = useLocation();

    const filtersDispatch = useFilterDispatch();
    const filtersState = useFilterState();

    const {initialized} = filtersState;
    
    const {systemFilters} = filtersState[systemFilterType || filterType];
    const {name: systemFilterName, suffix: systemSuffix, backPath: systemFilterBackPath, filter: systemFilter, customDisplay} = systemFilters;

    const {tableFilters, customFilters, selectedPageIndex, tableSort} = filtersState[filterType];

    const setTableFilters = (filters) => setFilters(filtersDispatch, {type: filterType, filters, isSystem: false});
    
    const resetSystemFilter = () => resetSystemFilters(filtersDispatch, systemFilterType || filterType);
    
    const fitlersList = [
        ...(!!filters ? [filters] : []),
        ...(!!tableFilters ? formatFiltersToOdataItems(tableFilters)  : []),
        ...(!!systemFilter ? [systemFilter]  : [])
    ];
    
    useEffect(() => {
        if (!initialized) {
            try {
                const {filterType, systemFilterType, tableFilters, systemFilters, customFilters} = JSON.parse(searchParams.get("filters") || {});
                
                initializeFilters(filtersDispatch, {filterType, systemFilterType, tableFilters, systemFilters, customFilters});
            } catch(error) {
                console.log("invalid filters");

                setSearchParams({}, { replace: true });
            }
        }
    }, [initialized, searchParams, filtersDispatch, setSearchParams]);

    useEffect(() => {
        const {customDisplay, ...cleanSystemFilters} = systemFilters;
        setSearchParams({filters: JSON.stringify({filterType, systemFilterType, tableFilters, systemFilters: cleanSystemFilters, customFilters})}, { replace: true });
    }, [filterType, systemFilterType, tableFilters, systemFilters, customFilters, setSearchParams]);

    if (!initialized) {
        return <Loader />;
    }
    
    return (
        <div style={!!withMargin && !!systemFilterName ? {marginTop: "80px"} : {}}>
            {!!systemFilterName &&
                <ErrorBoundary
                    FallbackComponent={({resetErrorBoundary}) => {
                        if (isEmpty(systemFilters)) {
                            resetErrorBoundary();
                        }
                        
                        return null;
                    }}
                    onError={resetSystemFilter}
                >
                    <SystemFilterBanner
                        displayText={<span>{`${toCapitalized(tableTitle)} for `}<BoldText>{systemFilterName}</BoldText>{` ${systemSuffix}`}</span>}
                        onClose={resetSystemFilter}
                        backPath={systemFilterBackPath}
                        customDisplay={customDisplay}
                    />
                </ErrorBoundary>
            }
            <div style={{...(!!withMargin ? {margin: "30px 30px 35px 30px"} : {marginBottom: "35px"}), position: "relative", visibility: (!!filtersConfig || !!CustomHeaderDisplay) ? "visible" : "hidden"}}>
                {!!filtersConfig &&
                    <ErrorBoundary
                        FallbackComponent={({resetErrorBoundary}) => {
                            if (isEmpty(tableFilters)) {
                                resetErrorBoundary();
                            }
                            
                            return null;
                        }}
                        onError={() => setFilters(filtersDispatch, {type: filterType, filters: [], isSystem: false})}
                    >
                        <Filter
                            filters={tableFilters}
                            onFilterUpdate={filters => setTableFilters(filters)}
                            filtersConfig={filtersConfig}
                            filtersOnCopyText={window.location.href}
                        />
                    </ErrorBoundary>
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
