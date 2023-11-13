import React, { useCallback, useEffect, useState } from 'react';

import OpenAPIParser from '@readme/openapi-parser';
import throttle from "lodash/throttle";
import { Utils as QbUtils, Query, Builder } from "@react-awesome-query-builder/ui";
import { useField } from 'formik';

import Button from 'components/Button';
//import openApiYaml from '/../api/openapi.yaml';
import { BASIC_CONFIG } from './CustomQueryBuilder.constants';
import { collectProperties } from './CustomQueryBuilder.functions';

//import openApiYaml from '../../../../../../api/openapi.yaml';

import "@react-awesome-query-builder/ui/css/styles.scss";

import "./CustomQueryBuilder.scss";

const openApiYaml = React.lazy(() => {
    console.log('process.env:', process.env);
    if (process.env.NODE_ENV !== "development") {
        return import('/src/src/openapi.yaml')
    }

    return import('/../api/openapi.yaml')
});

// You can load query value from your backend storage (for saving see `Query.onChange()`)
const queryValue = { id: QbUtils.uuid(), type: "group" };

const CustomQueryBuilder = ({
    initialQuery,
    name
}) => {
    const [config, setConfig] = useState(BASIC_CONFIG);
    const [queryState, setQueryState] = useState({
        config,
        tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config),
    });
    //process.env.NODE_ENV
    console.log('process.env.NODE_ENV:', process.env.NODE_ENV);
    const [field, , helpers] = useField(name);
    const { setValue } = helpers;
    const { value } = field;

    const readYamlFile = useCallback(
        async () => {
            try {
                const api = await OpenAPIParser.dereference(openApiYaml);
                //console.log("API name: %s, Version: %s", api.info.title, api.info.version);
                const properties = collectProperties(api.components.schemas.Asset);
                setConfig(previousConfig => ({ ...previousConfig, fields: properties }))
            } catch (err) {
                console.error(err);
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
            <div className="query-builder qb-lite">
                <Builder {...props} />
            </div>
        </div>
    ), []);

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
        setValue(query);
        // eslint-disable-next-line
    }, [queryState])


    useEffect(() => {
        setQueryState({ config, tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config) });
    }, [config])

    useEffect(() => {
        readYamlFile();
        // eslint-disable-next-line
    }, [])

    return (
        <div>
            <div className="query-builder-result">
                <div>
                    Human friendly query:{" "}
                    <pre className='query-builder-result__odata'>
                        {QbUtils.queryString(queryState.tree, queryState.config, true) ?? "-"}
                    </pre>
                </div>
                <div>
                    Query string:{" "}
                    <pre className='query-builder-result__odata'>
                        {value ?? "-"}
                    </pre>
                </div>
                {/* <div>
                    JsonLogic:{" "}
                    <pre>
                        {JSON.stringify(QbUtils.jsonLogicFormat(queryState.tree, queryState.config))}
                    </pre>
                </div> */}
            </div>
            <div className="query-buttons">
                <Button onClick={resetValue}>Reset</Button>
                <Button className="query-buttons__clear-button" onClick={clearValue}>Clear</Button>
            </div>
            <Query
                {...queryState.config}
                value={queryState.tree}
                onChange={onChange}
                renderBuilder={renderBuilder}
            />


        </div>
    )
}

export default CustomQueryBuilder;
