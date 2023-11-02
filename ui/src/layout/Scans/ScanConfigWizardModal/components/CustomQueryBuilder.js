import React, { useCallback, useEffect, useState } from 'react';
import { QueryBuilder, formatQuery } from 'react-querybuilder';
import { QueryBuilderMaterial } from '@react-querybuilder/material';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { teal } from '@mui/material/colors';
import { useField } from 'formik';
import openApiYaml from '../../../../../../api/openapi.yaml';
import { COMBINATORS, OPERATORS } from './CustomQueryBuilder.constants';
import OpenAPIParser from '@readme/openapi-parser';

const muiTheme = createTheme({
    palette: {
        secondary: {
            main: teal[500],
        },
    },
});

const collectProperties = (assetObject) => {
    const KEY_TO_FIND = "properties";
    const propertyList = {};
    const propertyListArr = [];


    const findProperties = (unknownValue) => {

        if (Array.isArray(unknownValue)) {
            unknownValue.forEach((subElement) => {
                findProperties(subElement);
            })
        } else if (typeof unknownValue === 'object') {
            for (const [key, value] of Object.entries(unknownValue)) {
                if (key === KEY_TO_FIND) {
                    Object.assign(propertyList, value);
                    //return true;
                } else {
                    //TODO: look for sub-properties
                    console.log(value);
                    findProperties(value);
                }
            }
        }
    }

    findProperties(assetObject);

    for (const [key, value] of Object.entries(propertyList)) {
        propertyListArr.push({
            ...value,
            name: key,
            label: key
        })
    }

    return propertyListArr;
}


const fields = [];

const CustomQueryBuilder = ({
    initialQuery,
    name
}) => {

    const [query, setQuery] = useState(initialQuery);
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
                const properties = collectProperties(api.components.schemas.Asset);
                console.log(properties);
            }
            catch (err) {
                console.error(err);
            }
        },
        [],
    );

    useEffect(() => {
        // convert query and setValue to scope
    }, [query])

    useEffect(() => {
        readYamlFile();
    }, [])

    return (
        <div>
            <ThemeProvider theme={muiTheme}>
                <QueryBuilderMaterial>
                    <QueryBuilder
                        operators={OPERATORS}
                        fields={fields}
                        query={query}
                        onQueryChange={q => setQuery(q)}
                    />
                </QueryBuilderMaterial>
            </ThemeProvider>
            <h4>Query</h4>
            <pre>
                <code>{formatQuery(query, 'json')}</code>
            </pre>
            {/* <TextField
                name="scanTemplate.scope"
                label="Scope"
                placeholder="Type an ODATA $filter to reduce assets to scan..."
            /> */}
        </div>
    )
}

export default CustomQueryBuilder;
