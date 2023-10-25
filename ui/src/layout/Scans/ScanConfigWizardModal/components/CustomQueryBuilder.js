import React, { useEffect, useState } from 'react';
import { QueryBuilder, formatQuery } from 'react-querybuilder';
import { QueryBuilderMaterial } from '@react-querybuilder/material';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { teal } from '@mui/material/colors';
import { useField } from 'formik';

const fs = require('fs-extra');

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

    useEffect(() => {
        // convert query and setValue to scope
    }, [query])

    useEffect(() => {
        // readYamlFile('../api/openapi.yml').then(data => {
        //     console.log(data)
        //     //=> {foo: true}
        // })
        const file = fs?.readFileSync('../api/openapi.yml', 'utf8');
        //console.log(YAML.parse(file));
        //console.log(data);

    }, [])

    return (
        <div>
            <ThemeProvider theme={muiTheme}>
                <QueryBuilderMaterial>
                    <QueryBuilder
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
