import React, { useMemo, useCallback, useEffect, useState } from 'react';
import classnames from 'classnames';
import { isEmpty, isEqual, pickBy, isNull } from 'lodash';
import { useTable, usePagination, useSortBy, useResizeColumns, useFlexLayout, useRowSelect } from 'react-table';
import Icon, { ICON_NAMES } from 'components/Icon';
import Loader from 'components/Loader';
import { useFetch, usePrevious } from 'hooks';
import Pagination from './Pagination';
import * as utils from './utils';

import './table.scss';

export {
    utils
}

const ACTIONS_COLUMN_ID = "ACTIONS";
const STATIC_COLUMN_IDS = [ACTIONS_COLUMN_ID];

const Table = props => {
    const {columns, defaultSortBy: defaultSortByItems, onLineClick, paginationItemsName, url, formatFetchedData, filters,
        noResultsTitle="items", refreshTimestamp, withPagination=true, data: externalData, onRowSelect,
        actionsComponent: ActionsComponent, customEmptyResultsDisplay: CustomEmptyResultsDisplay, actionsColumnWidth=80} = props;

    const defaultSortBy = useMemo(() => defaultSortByItems || [], [defaultSortByItems]);
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
        toggleSortBy, 
        state: {
            pageIndex,
            pageSize,
            sortBy,
            selectedRowIds
        }
    } = useTable({
            columns,
            getRowId: (rowData, rowIndex) => !!rowData.id ? rowData.id : rowIndex,
            data: tableItems,
            defaultColumn,
            initialState: {
                pageIndex: 0,
                pageSize: 50,
                sortBy: defaultSortBy,
                selectedRowIds: {}
            },
            manualPagination: true,
            pageCount: -1,
            manualSortBy: true,
            disableMultiSort: true
        },
        useResizeColumns,
        useFlexLayout,
        useSortBy,
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
                        disableSortBy: true,
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

    const {id: sortKey, desc: sortDesc} = !isEmpty(sortBy) ? sortBy[0] : {};
    const cleanFilters = pickBy(filters, value => !isNull(value) && value !== "");
    const prevCleanFilters = usePrevious(cleanFilters);
    const filtersChanged = !isEqual(cleanFilters, prevCleanFilters);
    const prevPageIndex = usePrevious(pageIndex);
    const prevSortKey = usePrevious(sortKey);
    const prevSortDesc = usePrevious(sortDesc);
    const sortingChanged = sortKey !== prevSortKey || sortDesc !== prevSortDesc;
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

        if (!isEmpty(sortKey)) {
            queryParams["$orderby"] = `${sortKey} ${sortDesc ? "desc" : "asc"}`;
        }

        return queryParams;
    }, [pageIndex, pageSize, sortKey, sortDesc, cleanFilters, withPagination]);

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
            gotoPage(0);

            return;
        }

        if (!!url) {
            doFetchWithQueryParams();
        }
    }, [filtersChanged, pageIndex, prevPageIndex, doFetchWithQueryParams, gotoPage, sortingChanged, refreshTimestamp, prevRefreshTimestamp, url]);

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

    if (isEmpty(page) && !loading && !!CustomEmptyResultsDisplay) {
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
                    gotoPage={gotoPage}
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
                                            const {id, isSorted, isSortedDesc} = column;

                                            return (
                                                <div className="table-th" {...column.getHeaderProps()}>
                                                    <span className="table-th-content">{column.render('Header')}</span>
                                                    {column.canSort && !column.disableSort &&
                                                        <Icon
                                                            className={classnames("table-sort-icon", {sorted: isSorted}, {rotate: isSortedDesc && isSorted})}
                                                            name={ICON_NAMES.SORT}
                                                            onClick={() => toggleSortBy(id, isSorted ? !isSortedDesc : false)}
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
                                    <div className={classnames("table-tr", {clickable: !!onLineClick}, {"with-row-actions": withRowActions})} {...row.getRowProps()}>
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
                                                    <div
                                                        className={cellClassName}
                                                        {...cell.getCellProps()}
                                                        onClick={() => {
                                                            if (!!onLineClick) {
                                                                onLineClick(row.original);
                                                            }
                                                        }}
                                                    >{isTextValue ? cell.value : cell.render('Cell')}</div>
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