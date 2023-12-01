import React, { useMemo, useCallback, useEffect, useState } from 'react';
import classnames from 'classnames';
import { isEmpty, isEqual, pickBy, isNull, isUndefined } from 'lodash';
import { useTable, usePagination, useResizeColumns, useFlexLayout, useRowSelect } from 'react-table';
import Icon, { ICON_NAMES } from 'components/Icon';
import Loader from 'components/Loader';
import { useFetch, usePrevious } from 'hooks';
import Pagination from './Pagination';
import * as utils from './utils';

import './table.scss';
import { ValueWithFallback } from 'components/ValueWithFallback';

export {
    utils
}

const ACTIONS_COLUMN_ID = "ACTIONS";
const STATIC_COLUMN_IDS = [ACTIONS_COLUMN_ID];

const Table = props => {
    const {columns, defaultSortBy, onSortChnage, onLineClick, paginationItemsName, url, formatFetchedData, filters, defaultPageIndex=0,
        onPageChange, noResultsTitle="items", refreshTimestamp, withPagination=true, data: externalData, onRowSelect,
        actionsComponent: ActionsComponent, customEmptyResultsDisplay: CustomEmptyResultsDisplay, showCustomEmptyDisplay=true,
        actionsColumnWidth=80} = props;

    const [sortBy, setSortBy] = useState(defaultSortBy || {});
    const prevSortBy = usePrevious(sortBy);

    useEffect(() => {
        if (!!onSortChnage && !isEqual(prevSortBy, sortBy)) {
            onSortChnage(sortBy);
        }
    }, [prevSortBy, sortBy, onSortChnage]);


    const defaultColumn = React.useMemo(() => ({
        minWidth: 30,
        width: 100
    }), []);

    const [{loading, data, error}, fetchData] = useFetch(url, {loadOnMount: false, formatFetchedData});
    const prevLoading = usePrevious(loading);
    const tableData = !!url ? data : externalData;
    const {items, count} = tableData || {};
    const tableItems = useMemo(() => items || [], [items]);

    const withRowActions = !!ActionsComponent;

    const [inititalLoaded, setInititalLoaded] = useState(!!externalData);

    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        prepareRow,
        page,
        canPreviousPage,
        nextPage,
        previousPage,
        gotoPage,
        state: {
            pageIndex,
            pageSize,
            selectedRowIds
        }
    } = useTable({
            columns,
            getRowId: (rowData, rowIndex) => !!rowData.id ? rowData.id : rowIndex,
            data: tableItems,
            defaultColumn,
            initialState: {
                pageIndex: defaultPageIndex,
                pageSize: 50,
                selectedRowIds: {}
            },
            manualPagination: true,
            pageCount: -1,
            disableMultiSort: true
        },
        useResizeColumns,
        useFlexLayout,
        usePagination,
        useRowSelect,
        hooks => {
            hooks.useInstanceBeforeDimensions.push(({headerGroups}) => {
                // fix the parent group of the framework columns to not be resizable
                headerGroups[0].headers.forEach(header => {
                    if (STATIC_COLUMN_IDS.includes(header.originalId)) {
                        header.canResize = false;
                    }
                });
            });

            hooks.visibleColumns.push(columns => {
                const updatedColumns = columns;
                
                if (withRowActions) {
                    updatedColumns.push({
                        Header: () => null, // No header
                        id: ACTIONS_COLUMN_ID,
                        accessor: original => (
                            <div className="actions-column-container">
                                {!!ActionsComponent && <ActionsComponent original={original} />}
                            </div>
                        ),
                        disableResizing: true,
                        minWidth: actionsColumnWidth,
                        width: actionsColumnWidth,
                        maxWidth: actionsColumnWidth
                    });
                }

                return updatedColumns;
            })
        }
    );

    const updatePage = useCallback(pageIndex => {
        if (!!onPageChange) {
            onPageChange(pageIndex);
        }

        gotoPage(pageIndex);
    }, [gotoPage, onPageChange]);

    const cleanFilters = pickBy(filters, value => !isNull(value) && value !== "");
    const prevCleanFilters = usePrevious(cleanFilters);
    const filtersChanged = !isEqual(cleanFilters, prevCleanFilters) && !isUndefined(prevCleanFilters);
    
    const prevPageIndex = usePrevious(pageIndex);

    const {sortIds: sortKeys, desc: sortDesc} = sortBy || {};
    const prevSortKeys = usePrevious(sortKeys);
    const prevSortDesc = usePrevious(sortDesc);
    const sortingChanged = !isEqual(sortKeys, prevSortKeys) || sortDesc !== prevSortDesc;

    const prevRefreshTimestamp = usePrevious(refreshTimestamp);

    const getQueryParams = useCallback(() => {
        const queryParams = {
            ...cleanFilters,
            "$count": true
        }

        if (withPagination) {
            queryParams["$skip"] = pageIndex * pageSize;
            queryParams["$top"] = pageSize;
        }

        if (!isEmpty(sortKeys)) {
            queryParams["$orderby"] = sortKeys.map(sortKey => `${sortKey} ${sortDesc ? "desc" : "asc"}`);
        }

        return queryParams;
    }, [pageIndex, pageSize, sortKeys, sortDesc, cleanFilters, withPagination]);

    const doFetchWithQueryParams = useCallback(() => {
        if (loading) {
            return;
        }
        
        fetchData({queryParams: {...getQueryParams()}});
    }, [fetchData, getQueryParams, loading])
    
    useEffect(() => {
        if (!filtersChanged && pageIndex === prevPageIndex && !sortingChanged && prevRefreshTimestamp === refreshTimestamp) {
            return;
        }

        if (filtersChanged && pageIndex !== 0) {
            updatePage(0);

            return;
        }

        if (!!url) {
            doFetchWithQueryParams();
        }
    }, [filtersChanged, pageIndex, prevPageIndex, doFetchWithQueryParams, updatePage, sortingChanged, refreshTimestamp, prevRefreshTimestamp, url]);

    const selectedRows = Object.keys(selectedRowIds);
    const prevSelectedRows = usePrevious(selectedRows);

    useEffect(() => {
        if (!!onRowSelect && !isEqual(selectedRows, prevSelectedRows)) {
            onRowSelect(selectedRows);
        }
    }, [prevSelectedRows, selectedRows, onRowSelect]);

    useEffect(() => {
        if (prevLoading && !loading && !inititalLoaded) {
            setInititalLoaded(true);
        }
    }, [prevLoading, loading, inititalLoaded]);

    if (!!error) {
        return null;
    }

    if (!inititalLoaded) {
        return (
            <Loader />
        )
    }

    if (isEmpty(page) && !loading && showCustomEmptyDisplay && !!CustomEmptyResultsDisplay) {
        return (
            <CustomEmptyResultsDisplay />
        )
    }
    
    return (
        <div className="table-wrapper">
            {!withPagination ? <div className="no-pagination-results-total">{`Showing ${count} entries`}</div> :
                <Pagination
                    canPreviousPage={canPreviousPage}
                    nextPage={nextPage}
                    previousPage={previousPage}
                    pageIndex={pageIndex}
                    pageSize={pageSize}
                    displayName={paginationItemsName}
                    gotoPage={updatePage}
                    loading={loading}
                    total={count}
                    page={page}
                />
            }
            <div className="table" {...getTableProps()}>
                <div className="table-head">
                    {
                        headerGroups.map(headerGroup => {
                            return (
                                <div className="table-tr" {...headerGroup.getHeaderGroupProps()}>
                                    {
                                        headerGroup.headers.map(column => {
                                            const {sortIds} = column;
                                            const isSorted = isEqual(sortIds, sortKeys);
                                            
                                            return (
                                                <div className="table-th" {...column.getHeaderProps()}>
                                                    <span className="table-th-content">{column.render('Header')}</span>
                                                    {!isEmpty(sortIds) &&
                                                        <Icon
                                                            className={classnames("table-sort-icon", {sorted: isSorted}, {rotate: isSorted && sortDesc})}
                                                            name={ICON_NAMES.SORT}
                                                            size={9}
                                                            onClick={() => setSortBy(({sortIds, desc}) => 
                                                                ({sortIds: column.sortIds, desc: isEqual(column.sortIds, sortIds) ? !desc : false})
                                                            )}
                                                        />
                                                    }
                                                    {column.canResize &&
                                                        <div
                                                            {...column.getResizerProps()}
                                                            className={classnames("resizer", {"isResizing": column.isResizing})}
                                                        />
                                                    }
                                                </div>
                                            )
                                        })
                                    }
                                </div>
                            )
                        })
                    }
                </div>
                <div className="table-body" {...getTableBodyProps()}>
                    {isEmpty(page) && !loading && <div className="empty-results-display-wrapper">{`No results available for ${noResultsTitle}`}</div>}
                    {loading && <div className="table-loading"><Loader /></div>}
                    {
                        page.map((row) => {
                            prepareRow(row);
                    
                            return (
                                <React.Fragment key={row.id}>
                                    <div
                                        className={classnames("table-tr", {clickable: !!onLineClick}, {"with-row-actions": withRowActions})}
                                        {...row.getRowProps()}
                                        onClick={() => {
                                            if (!!onLineClick) {
                                                onLineClick(row.original);
                                            }
                                        }}
                                    >
                                        {
                                            row.cells.map(cell => {
                                                const {className, alignToTop} = cell.column;
                                                const cellClassName = classnames(
                                                    "table-td",
                                                    {"align-to-top": alignToTop},
                                                    {[className]: className}
                                                );
                        
                                                const isTextValue = !!cell.column.accessor;
                                                
                                                return (
                                                    <div className={cellClassName} {...cell.getCellProps()}>
                                                        <ValueWithFallback>
                                                            {isTextValue ? cell.value : cell.render('Cell')}
                                                        </ValueWithFallback>
                                                    </div>
                                                )
                                            })
                                        }
                                    </div>
                                </React.Fragment>
                            )
                        })
                    }
                </div>
            </div>
            
        </div>
    )
}

export default React.memo(Table, (prevProps, nextProps) => {
    const {filters: prevFilters, refreshTimestamp: prevRefreshTimestamp, data: prevData} = prevProps;
    const {filters, refreshTimestamp, data} = nextProps;
    
    const shouldRefresh = !isEqual(prevFilters, filters) || prevRefreshTimestamp !== refreshTimestamp || !isEqual(prevData, data);
    
    return !shouldRefresh;
});
