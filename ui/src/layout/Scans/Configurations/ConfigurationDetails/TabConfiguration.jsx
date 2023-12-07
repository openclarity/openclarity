import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useFetch } from 'hooks';
import { TitleValueDisplayColumn } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Button from 'components/Button';
import Title from 'components/Title';
import Loader from 'components/Loader';
import ConfigurationReadOnlyDisplay from 'layout/Scans/ConfigurationReadOnlyDisplay';
import { ROUTES, APIS } from 'utils/systemConsts';
import { formatNumber } from 'utils/utils';
import { useFilterDispatch, setFilters, FILTER_TYPES } from 'context/FiltersProvider';

const ConfigurationScansDisplay = ({configId, configName}) => {
    const {pathname} = useLocation();
    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const scansFilter = `scanConfig/id eq '${configId}'`;

    const onScansClick = () => {
        setFilters(filtersDispatch, {
            type: FILTER_TYPES.SCANS,
            filters: {
                filter: scansFilter,
                name: configName,
                suffix: "configuration",
                backPath: pathname
            },
            isSystem: true
        });

        navigate(ROUTES.SCANS);
    }

    const [{loading, data, error}] = useFetch(APIS.SCANS, {queryParams: {"$filter": scansFilter, "$count": true}});
    
    if (error) {
        return null;
    }

    if (loading) {
        return <Loader absolute={false} small />;
    }
    
    return (
        <>
            <Title medium>Configuration's scans</Title>
            <Button onClick={onScansClick}>{`See all scans (${formatNumber(data?.count || 0)})`}</Button>
        </>
    )
}

const TabConfiguration = ({data}) => {
    const {id, name} = data || {};
    
    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <Title medium>Configuration</Title>
                    <TitleValueDisplayColumn>
                        <ConfigurationReadOnlyDisplay configData={data} />
                    </TitleValueDisplayColumn>
                </>
            )}
            rightPlaneDisplay={() => (
                <ConfigurationScansDisplay configId={id} configName={name} />
            )}
        />
    )
}

export default TabConfiguration;
