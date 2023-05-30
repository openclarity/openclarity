import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import { useSearchParams } from 'react-router-dom';
import { pickBy, isUndefined, isEmpty } from 'lodash';
import classnames from 'classnames';
import { ErrorBoundary } from 'react-error-boundary';
import Table from 'components/Table';
import SystemFilterDisplay from 'components/SystemFilterDisplay';
import Filter, { formatFiltersToQueryParams } from 'components/Filter';
import PageContainer from 'components/PageContainer';
import TopBarTitle from 'components/TopBarTitle';
import Loader from 'components/Loader';
import { useFilterState, useFilterDispatch, setFilters, setPage, resetSystemFilters, initializeFilters, setRuntimeScanFilter } from 'context/FiltersProvider';

import './table-page.scss';

export const TableHeaderPortal = ({children}) => {
    const [portalContainer, setPortalContainer] = useState(null);

    useEffect(() => {
        const container = document.querySelector(".table-header-wrapper");

        if (!container) {
            return;
        }
        
        setPortalContainer(container);
    }, []);

    if (!portalContainer) {
        return null;
    }

    return ReactDOM.createPortal(
        children,
        portalContainer
    );
}

const TablePage = ({columns, filterType, filtersMap, url, title, defaultSortBy, onLineClick, actionsComponent, refreshTimestamp: externalRefreshTimestamp}) => {
    const [searchParams, setSearchParams] = useSearchParams();

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const filtersState = useFilterState();
    const {initialized, currentRuntimeScan} = filtersState;
    const {tableFilters, systemFilters, selectedPageIndex} = filtersState[filterType];
    const filtersDispatch = useFilterDispatch();
    const setTableFilters = (filters) => setFilters(filtersDispatch, {type: filterType, filters, isSystem: false});
    
    const {title: systemFilterTitle, ...cleanSystemFilters} = systemFilters;

    const resetSystemFilter = () => resetSystemFilters(filtersDispatch, filterType);
    const resetRuntimeScanFilter = () => setRuntimeScanFilter(filtersDispatch, null);

    const hasSystemFitler = !!systemFilterTitle || !!currentRuntimeScan;
    
    useEffect(() => {
        doRefreshTimestamp();
    }, [externalRefreshTimestamp]);

    useEffect(() => {
        if (!initialized) {
            try {
                const {filterType, tableFilters, systemFilters, currentRuntimeScan} = JSON.parse(searchParams.get("filters") || {});

                initializeFilters(filtersDispatch, {filterType, tableFilters, systemFilters, currentRuntimeScan});
            } catch(error) {
                console.log("invalid filters");

                setSearchParams({});
            }
        }
    }, [initialized, searchParams, filtersDispatch, setSearchParams]);

    useEffect(() => {
        setSearchParams({filters: JSON.stringify({filterType, tableFilters, systemFilters, currentRuntimeScan})});
    }, [filterType, tableFilters, systemFilters, currentRuntimeScan, setSearchParams]);

    if (!initialized) {
        return <Loader />;
    }
    
    return (
        <div className="table-page-wrapper">
            <TopBarTitle title={title} onRefresh={doRefreshTimestamp} />
            <div className={classnames("table-page-content", {"with-padding": !hasSystemFitler})}>
                {!!systemFilterTitle &&
                    <ErrorBoundary
                        FallbackComponent={({resetErrorBoundary}) => {
                            if (isEmpty(systemFilters)) {
                                resetErrorBoundary();
                            }
                            
                            return null;
                        }}
                        onError={resetSystemFilter}
                    >
                        <SystemFilterDisplay displayText={<span>{`${title} for `}{systemFilterTitle}</span>} onClose={resetSystemFilter} />
                    </ErrorBoundary>
                }
                {!!currentRuntimeScan &&
                    <ErrorBoundary
                        FallbackComponent={({resetErrorBoundary}) => {
                            if (isEmpty(systemFilters)) {
                                resetErrorBoundary();
                            }
                            
                            return null;
                        }}
                        onError={resetRuntimeScanFilter}
                    >
                        <SystemFilterDisplay displayText={`${title} for the latest runtime scan`} onClose={resetRuntimeScanFilter} runtimeScanData={currentRuntimeScan} />
            
                    </ErrorBoundary>
                }
                <div className={classnames("table-wrapper", {"with-padding": !!hasSystemFitler})}>
                    <div className="table-header-wrapper">
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
                                filtersMap={filtersMap}
                                filtersOnCopyText={window.location.href}
                            />
                        </ErrorBoundary>
                    </div>
                    <PageContainer withPadding>
                        <Table
                            columns={columns}
                            paginationItemsName={title.toLowerCase()}
                            url={url}
                            defaultSortBy={defaultSortBy}
                            filters={{
                                ...pickBy(cleanSystemFilters, filter => !isUndefined(filter)),
                                ...formatFiltersToQueryParams(tableFilters),
                                ...(!!currentRuntimeScan ? {currentRuntimeScan: true} : {})
                            }}
                            onLineClick={onLineClick}
                            defaultPageIndex={selectedPageIndex}
                            onPageChange={pageIndex => setPage(filtersDispatch, {type: filterType, pageIndex})}
                            noResultsTitle={title}
                            refreshTimestamp={refreshTimestamp}
                            actionsComponent={actionsComponent}
                        />
                    </PageContainer>
                </div>
            </div>
        </div>
    )
}

export default TablePage;