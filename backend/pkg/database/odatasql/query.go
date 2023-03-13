// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package odatasql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/CiscoM31/godata"
)

var fixSelectToken sync.Once

// nolint:cyclop
func BuildCountQuery(schemaMetas map[string]SchemaMeta, schema string, filterString *string) (string, error) {
	// Parse top level $filter and create the top level "WHERE"
	var where string
	if filterString != nil && *filterString != "" {
		filterQuery, err := godata.ParseFilterString(context.TODO(), *filterString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $filter: %w", err)
		}

		// Build the WHERE conditions based on the $filter tree
		conditions, err := buildWhereFromFilter("Data", filterQuery.Tree)
		if err != nil {
			return "", fmt.Errorf("failed to build DB query from $filter: %w", err)
		}

		where = fmt.Sprintf("WHERE %s", conditions)
	}

	table := schemaMetas[schema].Table
	if table == "" {
		return "", fmt.Errorf("trying to query complex type schema %s with no source table", schema)
	}

	return fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, where), nil
}

// nolint:cyclop
func BuildSQLQuery(schemaMetas map[string]SchemaMeta, schema string, filterString, selectString, expandString *string, top, skip *int) (string, error) {
	// Fix GlobalExpandTokenizer so that it allows for `-` characters in the Literal tokens
	fixSelectToken.Do(func() {
		godata.GlobalExpandTokenizer.Add("^[a-zA-Z0-9_\\'\\.:\\$ \\*-]+", godata.ExpandTokenLiteral)
	})

	// Parse top level $filter and create the top level "WHERE"
	var where string
	if filterString != nil && *filterString != "" {
		filterQuery, err := godata.ParseFilterString(context.TODO(), *filterString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $filter: %w", err)
		}

		// Build the WHERE conditions based on the $filter tree
		conditions, err := buildWhereFromFilter("Data", filterQuery.Tree)
		if err != nil {
			return "", fmt.Errorf("failed to build DB query from $filter: %w", err)
		}

		where = fmt.Sprintf("WHERE %s", conditions)
	}

	var selectQuery *godata.GoDataSelectQuery
	if selectString != nil && *selectString != "" {
		// NOTE(sambetts):
		// For now we'll won't parse the data here and instead pass
		// just the raw value into the selectTree. The select tree will
		// parse the select query using the ExpandParser. If we can
		// update the GoData select parser to handle paths properly and
		// nest query params then we can switch to parsing select once
		// here before passing it to the selectTree.
		selectQuery = &godata.GoDataSelectQuery{RawValue: *selectString}
	}

	var expandQuery *godata.GoDataExpandQuery
	if expandString != nil && *expandString != "" {
		var err error
		expandQuery, err = godata.ParseExpandString(context.TODO(), *expandString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $expand ")
		}
	}

	// Turn the select and expand query params into a tree that can be used
	// to build nested select statements for the whole schema.
	//
	// TODO(sambetts) This should probably also validate that all the
	// selected/expanded fields are part of the schema.
	selectTree := newSelectTree()
	err := selectTree.insert(nil, nil, selectQuery, expandQuery, false)
	if err != nil {
		return "", fmt.Errorf("failed to parse select and expand: %w", err)
	}

	table := schemaMetas[schema].Table

	// Build query selecting fields based on the selectTree
	// For now all queries must start with a root "object" so we create a
	// complex field meta to represent that object
	rootObject := FieldMeta{FieldType: ComplexFieldType, ComplexFieldSchemas: []string{schema}}
	selectFields := buildSelectFields(schemaMetas, rootObject, schema, fmt.Sprintf("%s.Data", table), "$", selectTree)

	// Build paging statement
	var limitStm string
	if top != nil || skip != nil {
		limitVal := -1 // Negative means no limit, if no "$top" is specified this is what we want
		if top != nil {
			limitVal = *top
		}
		limitStm = fmt.Sprintf("LIMIT %d", limitVal)

		if skip != nil {
			limitStm = fmt.Sprintf("%s OFFSET %d", limitStm, *skip)
		}
	}

	if table == "" {
		return "", fmt.Errorf("trying to query complex type schema %s with no source table", schema)
	}

	return fmt.Sprintf("SELECT ID, %s AS Data FROM %s %s %s", selectFields, table, where, limitStm), nil
}

// nolint:cyclop,gocognit,gocyclo
func buildSelectFields(schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	switch field.FieldType {
	case PrimitiveFieldType:
		// If root of source (path is just $) is primitive just return the source
		if path == "$" {
			return source
		}
		return fmt.Sprintf("%s -> '%s'", source, path)
	case CollectionFieldType:
		newIdentifier := fmt.Sprintf("%sOptions", identifier)
		newSource := fmt.Sprintf("%s.value", identifier)

		var where string
		var newSelectNode *selectNode
		if st != nil {
			if st.filter != nil {
				conditions, _ := buildWhereFromFilter(newSource, st.filter.Tree)
				where = fmt.Sprintf("WHERE %s", conditions)
			}
			newSelectNode = &selectNode{children: st.children, expand: st.expand}
		}

		subQuery := buildSelectFields(schemaMetas, *field.CollectionItemMeta, newIdentifier, newSource, "$", newSelectNode)
		return fmt.Sprintf("(SELECT JSON_GROUP_ARRAY(%s) FROM JSON_EACH(%s, '%s') AS %s %s)", subQuery, source, path, identifier, where)
	case ComplexFieldType:
		// If there are no children in the select tree for this complex
		// type, shortcircuit and just return the data from the DB raw,
		// as there is no need to build the complex query, and it'll
		// ensure that null values are handled correctly.
		if st == nil || len(st.children) == 0 {
			return fmt.Sprintf("%s -> '%s'", source, path)
		}

		objects := []string{}
		for _, schemaName := range field.ComplexFieldSchemas {
			schema := schemaMetas[schemaName]

			parts := []string{}
			if field.DiscriminatorProperty != "" {
				parts = append(parts, fmt.Sprintf("'%s', '%s'", field.DiscriminatorProperty, schemaName))
			}
			for key, fm := range schema.Fields {
				if field.DiscriminatorProperty != "" && key == field.DiscriminatorProperty {
					continue
				}

				var sel *selectNode
				if st != nil {
					// If there are any select children
					// then we need to make sure this is
					// either a select child or a expand
					// child, otherwise skip this field.
					if len(st.selectChildren) > 0 {
						_, isSelect := st.selectChildren[key]
						_, isExpand := st.expandChildren[key]
						if !isSelect && !isExpand {
							continue
						}
					}
					sel = st.children[key]
				}

				extract := buildSelectFields(schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), source, fmt.Sprintf("%s.%s", path, key), sel)
				part := fmt.Sprintf("'%s', %s", key, extract)
				parts = append(parts, part)
			}
			objects = append(objects, fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(parts, ",")))
		}

		if len(objects) == 1 {
			return objects[0]
		}

		// TODO(sambetts) Error, if multiple schema there must be a
		// descriminator, this would be a developer error. Might be
		// avoidable if we create a schema builder thing instead of
		// just defining it as a variable.
		// if field.DiscriminatorProperty == "" {
		// }

		return fmt.Sprintf(
			"(SELECT %s.value FROM JSON_EACH(JSON_ARRAY(%s)) AS %s WHERE %s.value -> '$.%s' = %s -> '%s.%s')",
			identifier, strings.Join(objects, ","), identifier,
			identifier, field.DiscriminatorProperty, source, path, field.DiscriminatorProperty)

	case RelationshipFieldType:
		if st == nil || !st.expand {
			return fmt.Sprintf("%s -> '%s'", source, path)
		}

		schemaName := field.RelationshipSchema
		schema := schemaMetas[schemaName]
		newsource := fmt.Sprintf("%s.Data", schema.Table)
		parts := []string{}
		for key, fm := range schema.Fields {
			var sel *selectNode
			if st != nil {
				// If there are any select children
				// then we need to make sure this is
				// either a select child or a expand
				// child, otherwise skip this field.
				if len(st.selectChildren) > 0 {
					_, isSelect := st.selectChildren[key]
					_, isExpand := st.expandChildren[key]
					if !isSelect && !isExpand {
						continue
					}
				}
				sel = st.children[key]
			}

			extract := buildSelectFields(schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), newsource, fmt.Sprintf("$.%s", key), sel)
			part := fmt.Sprintf("'%s', %s", key, extract)
			parts = append(parts, part)
		}
		object := fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(parts, ","))

		return fmt.Sprintf("(SELECT %s FROM %s WHERE %s -> '$.%s' == %s -> '%s.%s')", object, schema.Table, newsource, field.RelationshipProperty, source, path, field.RelationshipProperty)
	case RelationshipCollectionFieldType:
		if st == nil || !st.expand {
			return fmt.Sprintf("%s -> '%s'", source, path)
		}

		schemaName := field.RelationshipSchema
		schema := schemaMetas[schemaName]
		newSource := fmt.Sprintf("%s.Data", schema.Table)

		where := fmt.Sprintf("WHERE %s -> '$.%s' = %s.value -> '$.%s'", newSource, field.RelationshipProperty, identifier, field.RelationshipProperty)
		if st != nil {
			if st.filter != nil {
				conditions, _ := buildWhereFromFilter(newSource, st.filter.Tree)
				where = fmt.Sprintf("%s and %s", where, conditions)
			}
		}

		parts := []string{}
		for key, fm := range schema.Fields {
			var sel *selectNode
			if st != nil {
				// If there are any select children
				// then we need to make sure this is
				// either a select child or a expand
				// child, otherwise skip this field.
				if len(st.selectChildren) > 0 {
					_, isSelect := st.selectChildren[key]
					_, isExpand := st.expandChildren[key]
					if !isSelect && !isExpand {
						continue
					}
				}
				sel = st.children[key]
			}

			extract := buildSelectFields(schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), newSource, fmt.Sprintf("$.%s", key), sel)
			part := fmt.Sprintf("'%s', %s", key, extract)
			parts = append(parts, part)
		}
		subQuery := fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(parts, ","))

		return fmt.Sprintf("(SELECT JSON_GROUP_ARRAY(%s) FROM %s,JSON_EACH(%s, '%s') AS %s %s)", subQuery, schema.Table, source, path, identifier, where)
	default:
		return ""
	}
}

var sqlOperators = map[string]string{
	"eq":         "=",
	"ne":         "!=",
	"gt":         ">",
	"ge":         ">=",
	"lt":         "<",
	"le":         "<=",
	"or":         "or",
	"contains":   "%%%s%%",
	"endswith":   "%%%s",
	"startswith": "%s%%",
}

func singleQuote(s string) string {
	return fmt.Sprintf("'%s'", s)
}

func buildJSONPathFromParseNode(node *godata.ParseNode) (string, error) {
	switch node.Token.Type {
	case godata.ExpressionTokenNav:
		right, err := buildJSONPathFromParseNode(node.Children[0])
		if err != nil {
			return "", fmt.Errorf("unable to build right side of navigation path: %w", err)
		}

		left, err := buildJSONPathFromParseNode(node.Children[1])
		if err != nil {
			return "", fmt.Errorf("unable to build left side of navigation path: %w", err)
		}
		return fmt.Sprintf("%s.%s", right, left), nil
	case godata.ExpressionTokenLiteral:
		return node.Token.Value, nil
	default:
		return "", fmt.Errorf("unsupported token type")
	}
}

// TODO: create a unit test
// nolint:cyclop
func buildWhereFromFilter(source string, node *godata.ParseNode) (string, error) {
	operator := node.Token.Value

	var query string
	switch operator {
	case "eq", "ne", "gt", "ge", "lt", "le":
		// Convert ODATA paths with slashes like "Thing/Name" into JSON
		// path like "Thing.Name".
		queryPath, err := buildJSONPathFromParseNode(node.Children[0])
		if err != nil {
			return "", fmt.Errorf("unable to covert oData path to json path: %w", err)
		}

		rhs := node.Children[1]
		extractFunction := "->"
		sqlOperator := sqlOperators[operator]
		var value string
		switch rhs.Token.Type { // TODO: implement all the relevant cases as ExpressionTokenDate and ExpressionTokenDateTime
		case godata.ExpressionTokenString:
			value = singleQuote(strings.ReplaceAll(rhs.Token.Value, "'", "\""))
		case godata.ExpressionTokenBoolean:
			value = singleQuote(rhs.Token.Value)
		case godata.ExpressionTokenInteger, godata.ExpressionTokenFloat:
			value = rhs.Token.Value
			extractFunction = "->>"
		case godata.ExpressionTokenNull:
			value = "NULL"
			if operator == "eq" {
				sqlOperator = "is"
			} else if operator == "ne" {
				sqlOperator = "is not"
			} else {
				return "", fmt.Errorf("unsupported ExpressionTokenNull operator %s", operator)
			}
		default:
			return "", fmt.Errorf("unsupported token type %s", node.Children[1].Token.Type)
		}

		query = fmt.Sprintf("%s %s '%s' %s %s", source, extractFunction, queryPath, sqlOperator, value)
	case "and":
		left, err := buildWhereFromFilter(source, node.Children[0])
		if err != nil {
			return query, err
		}
		right, err := buildWhereFromFilter(source, node.Children[1])
		if err != nil {
			return query, err
		}
		query = fmt.Sprintf("(%s AND %s)", left, right)
	case "or":
		left, err := buildWhereFromFilter(source, node.Children[0])
		if err != nil {
			return query, err
		}
		right, err := buildWhereFromFilter(source, node.Children[1])
		if err != nil {
			return query, err
		}
		query = fmt.Sprintf("(%s OR %s)", left, right)
	case "contains", "endswith", "startswith":
		queryField := node.Children[0].Token.Value
		queryPath := fmt.Sprintf("$.%s", queryField)

		right := node.Children[1].Token.Value
		var value interface{}
		switch node.Children[1].Token.Type {
		case godata.ExpressionTokenString:
			r := strings.ReplaceAll(right, "'", "")
			value = fmt.Sprintf(sqlOperators[operator], r)
		default:
			return query, fmt.Errorf("unsupported token type")
		}
		query = fmt.Sprintf("%s -> '%s' LIKE '%s'", source, queryPath, value)
	}

	return query, nil
}
