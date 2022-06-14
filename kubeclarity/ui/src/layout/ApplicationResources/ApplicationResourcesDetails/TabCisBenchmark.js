import React, { useMemo } from 'react';
import Table from 'components/Table';
import Icon, { ICON_NAMES } from 'components/Icon';
import { CIS_SEVERITY_ITEMS } from 'utils/systemConsts';

const TabCisBenchmark = ({id}) => {
    const columns = useMemo(() => [
        {
            Header: "Name",
            id: "code",
            accessor: "code"
        },
        {
            Header: "Title",
            id: "title",
            accessor: "title",
            disableSort: true
        },
        {
            Header: "Findings",
            id: "level",
            Cell: ({row}) => {
                const {level} = row.original;
                const {label, color} = CIS_SEVERITY_ITEMS[level] || {};
                
                return (
                    <div className="cis-findings-level-icon">
                        <Icon name={ICON_NAMES.ALERT} style={{color}} />
                        <span style={{color}}>{label}</span>
                    </div>
                )
            },
            canSort: true
        },
        {
            Header: "Description",
            id: "desc",
            accessor: "desc",
            disableSort: true
        }
    ], []);

    return (
        <div className="application-resource-tab-cis-benchmark">
            <Table
                columns={columns}
                url={`cisdockerbenchmarkresults/${id}`}
                defaultSortBy={[{id: "code", desc: false}]}
                withPagination={false}
                noResultsTitle="CIS Benchmark"
            />
        </div>
    )
}

export default TabCisBenchmark;
