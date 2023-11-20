import OpenAPIParser from '@readme/openapi-parser';
import React, { useCallback, useEffect, useState } from 'react';
import classNames from "classnames";
import throttle from "lodash/throttle";
import { Query, Builder, Utils as QbUtils } from '@react-awesome-query-builder/mui';
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

// You can load query value from your backend storage (for saving see `Query.onChange()`)
const queryValue = { id: QbUtils.uuid(), type: "group" };

const CustomQueryBuilder = ({
    initialQuery,
    name
}) => {
    const isDev = process.env.NODE_ENV === "development";
    const [{ loading, data, error }] = useFetch(`${isDev ? "http://localhost:3000" : ""}/api/openapi.json`, { isAbsoluteUrl: true });
    const [config, setConfig] = useState(BASIC_CONFIG);
    const [queryState, setQueryState] = useState({
        config,
        tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config),
    });

    const [, , helpers] = useField(name); //const [field, , helpers] 
    const { setValue } = helpers;
    //const { value } = field;

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
            tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config),
        }));
    }, [config]);

    const clearValue = useCallback(() => {
        setQueryState(state => ({
            ...state,
            tree: QbUtils.loadTree(queryValue),
        }));
    }, []);

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
        //const jsonTree = QbUtils.getTree(immutableTree);
        //console.log(jsonTree);
        // `jsonTree` can be saved to backend, and later loaded to `queryValue`
        updateResult(immutableTree, config)
    }, [updateResult]);

    useEffect(() => {
        const query = QbUtils.queryString(queryState.tree, queryState.config);
        setValue(postFixQuery(query));
        // eslint-disable-next-line
    }, [queryState])


    useEffect(() => {
        setQueryState({ config, tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config) });
    }, [config])

    useEffect(() => {
        readYamlFile(data);
        // eslint-disable-next-line
    }, [data])

    return (
        <div>
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
        </div>
    )
}

export default CustomQueryBuilder;
