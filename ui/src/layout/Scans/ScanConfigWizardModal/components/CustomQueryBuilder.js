import React, { useCallback, useEffect, useState } from 'react';
import { QueryBuilder, formatQuery } from 'react-querybuilder';
import { QueryBuilderMaterial } from '@react-querybuilder/material';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { teal } from '@mui/material/colors';
import { useField } from 'formik';
import openApiYaml from '../../../../../../api/openapi.yaml';
import { COMBINATORS, OPERATORS } from './CustomQueryBuilder.constants';

const muiTheme = createTheme({
    palette: {
        secondary: {
            main: teal[500],
        },
    },
});

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
        () => {
            const assets = openApiYaml?.components?.schemas?.Asset
            console.log(assets);
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
