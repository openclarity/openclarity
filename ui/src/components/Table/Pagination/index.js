import React from 'react';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';

import './pagination.scss';

const FIRST_PAGE_NUMBER = 1;

const PaginationItem = ({children, className, isActive=false, onClick}) => (
    <div className={classnames("pagination-item", className, {"is-active": isActive})} onClick={onClick}>{children}</div>
);

const PaginationNumber = ({page, isActive, onClick}) => (
    <PaginationItem onClick={onClick} isActive={isActive}>{String(page)}</PaginationItem>
);

const PaginationArrow = ({isLeft=false, isDouble=false, onClick, disabled}) => (
    <PaginationItem className={classnames("pagination-arrow", {left: isLeft}, {disabled})} onClick={disabled ? undefined : onClick}>
        <Icon name={isDouble ? ICON_NAMES.CHEVRON_RIGHT_DOUBLE : ICON_NAMES.CHEVRON_RIGHT} size={19} />
    </PaginationItem>
);

const PaginationDots = () => (
    <PaginationItem className="pagination-dots">...</PaginationItem>
);

const Pagination = ({canPreviousPage, previousPage, nextPage, pageIndex, pageSize, gotoPage, page, total, loading, displayName="items"}) => {
    if (total === 0 || loading) {
        return <div className="pagination-container"></div>;
    }
    
    const startPageItem = (pageIndex * pageSize) + 1;
    const endPageItem = page.length + (pageIndex * pageSize);
    const lastPageNumber = Math.ceil(total / pageSize) || 1;
    const activePageNumber = pageIndex + 1;
    const canNextPage = lastPageNumber > activePageNumber;
    
    const prevPageNumber = activePageNumber - 1;
    const nextPageNumber = activePageNumber + 1;
    const displayPageNumbers = [...new Set([
        prevPageNumber < FIRST_PAGE_NUMBER ? FIRST_PAGE_NUMBER : prevPageNumber,
        activePageNumber,
        nextPageNumber > lastPageNumber ? lastPageNumber : nextPageNumber
    ])].filter(page => page !== FIRST_PAGE_NUMBER && page !== lastPageNumber);
    
    const goToPageNumber = page => gotoPage(page - 1);
    const goToFirstPage = () => goToPageNumber(FIRST_PAGE_NUMBER);
    const goToLastPage = () => goToPageNumber(lastPageNumber);

    return (
        <div className="pagination-container">
            <div className="pagination-results">
                {`Showing ${startPageItem}-${endPageItem} of ${total} ${displayName}`}
            </div>
            <div className="pagination-navigation">
                <PaginationArrow isLeft isDouble onClick={goToFirstPage} disabled={!canPreviousPage} />
                <PaginationArrow isLeft onClick={previousPage} disabled={!canPreviousPage} />
                <PaginationNumber page={FIRST_PAGE_NUMBER} isActive={activePageNumber === FIRST_PAGE_NUMBER} onClick={goToFirstPage} />
                {displayPageNumbers[0] > (FIRST_PAGE_NUMBER + 1) && <PaginationDots />}
                {
                    displayPageNumbers.map(page => (
                        <PaginationNumber key={page} page={page} onClick={() => goToPageNumber(page)} isActive={page === activePageNumber} />
                    ))
                }
                {displayPageNumbers[displayPageNumbers.length - 1] < (lastPageNumber - 1) && <PaginationDots />}
                {FIRST_PAGE_NUMBER !== lastPageNumber &&
                    <PaginationNumber page={lastPageNumber} isActive={activePageNumber === lastPageNumber} onClick={goToLastPage} />
                }
                <PaginationArrow onClick={nextPage} disabled={!canNextPage} />
                <PaginationArrow isDouble onClick={goToLastPage} disabled={!canNextPage} />
            </div>
        </div>
    );
}

export default Pagination;