import OpenAPIParser from '@readme/openapi-parser';
import React, { useCallback, useEffect, useState } from 'react';
import classNames from "classnames";
import throttle from "lodash/throttle";
import { Utils as QbUtils } from '@react-awesome-query-builder/core';
import { Query, Builder } from '@react-awesome-query-builder/mui';
import { useFetch } from 'hooks';
import { useField } from 'formik';

import Button from 'components/Button';
import FieldError from 'components/Form/FieldError';
import Loader from 'components/Loader';
import { BASIC_CONFIG } from './CustomQueryBuilder.constants';
import { TextAreaField } from 'components/Form';
import { collectProperties, postFixQuery } from './CustomQueryBuilder.functions';

import "@react-awesome-query-builder/ui/css/styles.scss";
import "./CustomQueryBuilder.scss";

const CustomQueryBuilder = () => {
    const [{ loading, data, error }] = useFetch(`${window.location.origin}/api/openapi.json`, { isAbsoluteUrl: true });
    /* BASIC_CONFIG is the configuration object for the query builder, 'scopeConfig' contains all the tree and query data -- this one is saved to the backend */

    const [config, setConfig] = useState(BASIC_CONFIG);

    const [scopeConfigField, , scopeConfigHelpers] = useField("scanTemplate.scopeConfig");
    const { setValue: setScopeConfigValue } = scopeConfigHelpers;
    const { value: scopeConfigValue } = scopeConfigField;

    const [, , scopeHelpers] = useField("scanTemplate.scope");
    const { setValue: setScopeValue } = scopeHelpers;

    const [queryState, setQueryState] = useState({
        config,
        tree: QbUtils.checkTree(QbUtils.loadTree(scopeConfigValue), config),
    });

    const readYamlFile = useCallback(
        async (rawApiData) => {
            if (rawApiData) {
                try {
                    const apiData = await OpenAPIParser.dereference(rawApiData);
                    const properties = collectProperties(apiData.components.schemas.Asset);
                    setConfig(previousConfig => ({ ...previousConfig, fields: properties }))
                } catch (err) {
                    console.error(err);
                }
            }
        },
        [],
    );

    const resetValue = useCallback(() => {
        setQueryState(state => ({
            ...state,
            tree: QbUtils.checkTree(QbUtils.loadTree(scopeConfigValue), config),
        }));
    }, [config, scopeConfigValue]);

    const clearValue = useCallback(() => {
        setQueryState(state => ({
            ...state,
            tree: QbUtils.loadTree(scopeConfigValue),
        }));
    }, [scopeConfigValue]);

    const renderBuilder = useCallback((props) => (
        <div className="query-builder-container">
            <div className={classNames("query-builder", { "qb-lite": queryState.tree.size > 2 })}>
                <Builder {...props} />
            </div>
        </div>
    ), [queryState.tree.size]);

    const updateResult = throttle((immutableTree, config) => {
        setQueryState(prevState => ({ ...prevState, tree: immutableTree, config }));
        // eslint-disable-next-line
    }, 100);

    const onChange = useCallback((immutableTree, config) => {
        updateResult(immutableTree, config)
        const jsonTree = QbUtils.getTree(immutableTree);
        setScopeConfigValue(jsonTree);
    }, [setScopeConfigValue, updateResult]);

    useEffect(() => {
        const query = QbUtils.queryString(queryState.tree, queryState.config);
        setScopeValue(postFixQuery(query));
        // eslint-disable-next-line
    }, [queryState.tree])

    useEffect(() => {
        setQueryState({
            config,
            tree: QbUtils.checkTree(QbUtils.loadTree(scopeConfigValue), config)
        });
        // eslint-disable-next-line
    }, [config])

    useEffect(() => {
        readYamlFile(data);
        // eslint-disable-next-line
    }, [data])

    return (
        <>
            <div className="query-builder-result">
                <div className='query-builder-result__section'>
                    <span className='query-builder-result__title'>Manual scope editor (odata query)*</span>
                    <div className='query-builder-result__odata'>
                        <span className='query-builder-result__odata--details'>(This query is going to be used by the scanner)</span>
                        <TextAreaField
                            name="scanTemplate.scope"
                            placeholder="You can type a scope manually..."
                        />
                    </div>
                </div>
                <div className='query-builder-result__section'>
                    <span className='query-builder-result__title'>Human friendly scope:{" "}</span>
                    <div className='query-builder-result__odata'>
                        {QbUtils.queryString(queryState.tree, queryState.config, true) ?? "-"}
                    </div>
                </div>
                <div className="query-buttons">
                    <Button onClick={resetValue}>Reset</Button>
                    <Button className="query-buttons__clear-button" onClick={clearValue}>Clear</Button>
                </div>
            </div>

            {loading && <Loader absolute={false} />}
            {error && <FieldError>{error?.message}</FieldError>}
            {Object.keys(config.fields).length > 0 &&
                <>
                    <Query
                        {...queryState.config}
                        value={queryState.tree}
                        onChange={onChange}
                        renderBuilder={renderBuilder}
                    />
                </>
            }
        </>
    )
}

export default CustomQueryBuilder;
