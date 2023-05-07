import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import { pickBy, isUndefined } from 'lodash';
import classnames from 'classnames';
import Table from 'components/Table';
import SystemFilterDisplay from 'components/SystemFilterDisplay';
import Filter, { formatFiltersToQueryParams } from 'components/Filter';
import PageContainer from 'components/PageContainer';
import TopBarTitle from 'components/TopBarTitle';
import { useFilterState, useFilterDispatch, setFilters, setPage, resetSystemFilters } from 'context/FiltersProvider';

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
    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    const filtersState = useFilterState();
    const {tableFilters, systemFilters, selectedPageIndex} = filtersState[filterType];
    const filtersDispatch = useFilterDispatch();
    const setTableFilters = (filters) => setFilters(filtersDispatch, {type: filterType, filters, isSystem: false});
    
    const {title: systemFilterTitle, currentRuntimeScan, ...cleanSystemFilters} = systemFilters;

    const onSystemFilterClose = () => resetSystemFilters(filtersDispatch, filterType);

    const hasSystemFitler = !!systemFilterTitle || !!currentRuntimeScan;
    
    useEffect(() => {
        doRefreshTimestamp();
    }, [externalRefreshTimestamp]);
    
    return (
        <div className="table-page-wrapper">
            <TopBarTitle title={title} onRefresh={doRefreshTimestamp} />
            <div className={classnames("table-page-content", {"with-padding": !hasSystemFitler})}>
                {!!systemFilterTitle && <SystemFilterDisplay displayText={<span>{`${title} for `}{systemFilterTitle}</span>} onClose={onSystemFilterClose} />}
                {!!currentRuntimeScan && <SystemFilterDisplay displayText={`${title} for the latest runtime scan`} onClose={onSystemFilterClose} runtimeScanData={currentRuntimeScan} />}
                <div className={classnames("table-wrapper", {"with-padding": !!hasSystemFitler})}>
                    <div className="table-header-wrapper">
                        <Filter
                            filters={tableFilters}
                            onFilterUpdate={filters => setTableFilters(filters)}
                            filtersMap={filtersMap}
                        />
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