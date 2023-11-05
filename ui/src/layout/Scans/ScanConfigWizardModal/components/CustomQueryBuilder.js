import React, { useCallback, useEffect, useState } from 'react';

import OpenAPIParser from '@readme/openapi-parser';
import cloneDeep from "lodash/cloneDeep";
import throttle from "lodash/throttle";
import { Utils as QbUtils, Query, Builder } from "@react-awesome-query-builder/ui";
import { useField } from 'formik';

import Button from 'components/Button';
import openApiYaml from '../../../../../../api/openapi.yaml';
import { BASIC_CONFIG } from './CustomQueryBuilder.constants';
import { convertDataType, postConvertQuery } from './CustomQueryBuilder.functions';

import "@react-awesome-query-builder/ui/css/styles.scss";

const collectProperties = (assetObject) => {
    console.log('assetObject:', cloneDeep(assetObject));

    const KEY_TO_FIND = "properties";
    const propertyList = {};

    const prepareProperties = (unknownValue, targetObject) => {
        if (Array.isArray(unknownValue)) {
            unknownValue.forEach((subElement) => {
                prepareProperties(subElement, targetObject);
            })
        } else if (typeof unknownValue === 'object') {
            for (const [key, value] of Object.entries(unknownValue)) {
                if (key === KEY_TO_FIND && typeof value === 'object' && !Array.isArray(value)) {
                    Object.assign(targetObject, value)
                } else {
                    prepareProperties(value, targetObject);
                }
            }
        }
    }

    const createSubfields = (source, targetObject) => {
        if (Array.isArray(source)) {
            source.forEach((subElement) => {
                createSubfields(subElement, targetObject);
            })
        } else if (typeof source === 'object') {
            for (const [key, value] of Object.entries(source)) {
                if (key === KEY_TO_FIND && typeof value === 'object' && !Array.isArray(value)) {
                    //value here is always the properties object
                    if (targetObject.hasOwnProperty("subfields")) {
                        Object.assign(targetObject.subfields, value)
                    } else {
                        targetObject.subfields = value;
                    }
                    for (const [subKey, subValue] of Object.entries(value)) {
                        createSubfields(subValue, targetObject.subfields[subKey]);
                    }
                } else {
                    createSubfields(value, targetObject);
                }
            }
        }
    }

    const convertDataForQuery = (propertyList) => {
        if (Array.isArray(propertyList)) {
            propertyList.forEach((subElement) => {
                convertDataForQuery(subElement);
            })
        } else if (typeof propertyList === 'object') {
            for (const [key, value] of Object.entries(propertyList)) {
                if (Object.keys(value).includes("type")) {
                    value.type = convertDataType(value.type);
                    value.label = key;

                    if (value.enum) {
                        value.type = 'select';
                        value.fieldSettings = {
                            listValues: value.enum.map(opt => ({
                                title: opt,
                                value: opt
                            }))
                        }
                    }

                    if (value.type === 'text' && value.format === 'date-time') {
                        value.type = 'datetime';
                    }
                }

                if (Object.keys(value).includes("subfields")) {
                    propertyList[`${key}-object`] = {
                        type: 'group-select',
                        label: `${key} (object)`,
                        fieldSettings: {
                            listValues: [{
                                title: 'null',
                                value: 'null'
                            }]
                        },
                        fieldName: key,
                        defaultValue: "null",
                    }
                }

                convertDataForQuery(value);
            }
        }
    }

    prepareProperties(cloneDeep(assetObject), propertyList);
    Object.values(propertyList).forEach(value => createSubfields(value, value))
    convertDataForQuery(propertyList);

    console.log('propertyList:', propertyList);

    return propertyList;
}


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
    //const [query, setQuery] = useState(initialQuery);
    const [field, helpers] = useField(name);
    const { value } = field;
    const { setValue } = helpers;

    const readYamlFile = useCallback(
        async () => {
            //const assets = openApiYaml?.components?.schemas?.Asset

            try {
                let api = await OpenAPIParser.dereference(openApiYaml);
                //const api = OpenAPIParser.dereference(openApiYaml)

                console.log("API name: %s, Version: %s", api.info.title, api.info.version);
                console.log(api.components.schemas.Asset);
                console.log('bc: ', BASIC_CONFIG)
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
        <div className="query-builder-container" style={{ padding: "10px" }}>
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
        // Tip: for better performance you can apply `throttle` - see `examples/demo`
        //setQueryState(prevState => ({ ...prevState, tree: immutableTree, config: config }));
        updateResult(immutableTree, config)
        const jsonTree = QbUtils.getTree(immutableTree);
        console.log(jsonTree);
        // `jsonTree` can be saved to backend, and later loaded to `queryValue`
        // TODO: SAVE TO BACKEND
    }, [updateResult]);

    useEffect(() => {
        setQueryState({ config, tree: QbUtils.checkTree(QbUtils.loadTree(queryValue), config) });
    }, [config])

    useEffect(() => {
        readYamlFile();
        // eslint-disable-next-line
    }, [])

    return (
        <div>
            <Query
                {...queryState.config}
                value={queryState.tree}
                onChange={onChange}
                renderBuilder={renderBuilder}
            />
            <div className="query-builder-result">
                <div>
                    Query string:{" "}
                    <pre>
                        {JSON.stringify(postConvertQuery(QbUtils.queryString(queryState.tree, queryState.config)))}
                    </pre>
                </div>
                <div>
                    JsonLogic:{" "}
                    <pre>
                        {JSON.stringify(QbUtils.jsonLogicFormat(queryState.tree, queryState.config))}
                    </pre>
                </div>
            </div>

            <Button onClick={resetValue}>Reset</Button>
            <Button onClick={clearValue}>Clear</Button>

        </div>
    )
}

export default CustomQueryBuilder;
