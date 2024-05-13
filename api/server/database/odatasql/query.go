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

// nolint:perfsprint
package odatasql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/CiscoM31/godata"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/server/database/odatasql/jsonsql"
)

var fixSelectToken sync.Once

// nolint:cyclop
func BuildCountQuery(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, schema string, filterString *string) (string, error) {
	table := schemaMetas[schema].Table
	if table == "" {
		return "", fmt.Errorf("trying to query complex type schema %s with no source table", schema)
	}

	// Build query selecting fields based on the selectTree
	// For now all queries must start with a root "object" so we create a
	// complex field meta to represent that object
	rootObject := FieldMeta{FieldType: ComplexFieldType, ComplexFieldSchemas: []string{schema}}

	// Parse top level $filter and create the top level "WHERE"
	var where string
	if filterString != nil && *filterString != "" {
		filterQuery, err := godata.ParseFilterString(context.TODO(), *filterString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $filter: %w", err)
		}

		// Build the WHERE conditions based on the $filter tree
		conditions, err := buildWhereFromFilter(sqlVariant, schemaMetas, rootObject, schema, fmt.Sprintf("%s.Data", table), filterQuery.Tree, "")
		if err != nil {
			return "", fmt.Errorf("failed to build DB query from $filter: %w", err)
		}

		where = fmt.Sprintf("WHERE %s", conditions)
	}

	return fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, where), nil
}

// nolint:cyclop,gocognit
func BuildSQLQuery(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, schema string, filterString, selectString, expandString, orderbyString *string, top, skip *int) (string, error) {
	// Fix GlobalExpandTokenizer so that it allows for `-` characters in the Literal tokens
	fixSelectToken.Do(func() {
		godata.GlobalExpandTokenizer.Add("^[a-zA-Z0-9_\\'\\.:\\$ \\*-]+", godata.ExpandTokenLiteral)
	})

	table := schemaMetas[schema].Table
	if table == "" {
		return "", fmt.Errorf("trying to query complex type schema %s with no source table", schema)
	}

	// Build query selecting fields based on the selectTree
	// For now all queries must start with a root "object" so we create a
	// complex field meta to represent that object
	rootObject := FieldMeta{FieldType: ComplexFieldType, ComplexFieldSchemas: []string{schema}}

	// Parse top level $filter and create the top level "WHERE"
	var where string
	if filterString != nil && *filterString != "" {
		filterQuery, err := godata.ParseFilterString(context.TODO(), *filterString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $filter: %w", err)
		}

		// Build the WHERE conditions based on the $filter tree
		conditions, err := buildWhereFromFilter(sqlVariant, schemaMetas, rootObject, schema, fmt.Sprintf("%s.Data", table), filterQuery.Tree, "")
		if err != nil {
			return "", fmt.Errorf("failed to build DB query from $filter: %w", err)
		}

		where = fmt.Sprintf("WHERE %s", conditions)
	}

	var orderby string
	if orderbyString != nil && *orderbyString != "" {
		orderbyQuery, err := godata.ParseOrderByString(context.TODO(), *orderbyString)
		if err != nil {
			return "", fmt.Errorf("failed to parse $orderby: %w", err)
		}

		conditions, err := buildOrderByFromOdata(sqlVariant, schemaMetas, rootObject, schema, fmt.Sprintf("%s.Data", table), orderbyQuery.OrderByItems)
		if err != nil {
			return "", fmt.Errorf("failed to build DB query from $orderby: %w", err)
		}

		orderby = fmt.Sprintf("ORDER BY %s", conditions)
	}

	selectFields, err := buildSelectFieldsFromSelectAndExpand(sqlVariant, schemaMetas, rootObject, schema, fmt.Sprintf("%s.Data", table), selectString, expandString)
	if err != nil {
		return "", fmt.Errorf("failed to construct fields to select: %w", err)
	}

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

	return fmt.Sprintf("SELECT ID, %s AS Data FROM %s %s %s %s", selectFields, table, where, orderby, limitStm), nil
}

func buildSelectFieldsFromSelectAndExpand(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, rootObject FieldMeta, identifier string, source string, selectString, expandString *string) (string, error) {
	var selectQuery *godata.GoDataSelectQuery
	if selectString != nil && *selectString != "" {
		// NOTE(sambetts):
		// For now we won't parse the data here and instead pass
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
	err := selectTree.insert(nil, nil, nil, selectQuery, expandQuery, false)
	if err != nil {
		return "", fmt.Errorf("failed to parse select and expand: %w", err)
	}

	return buildSelectFields(sqlVariant, schemaMetas, rootObject, identifier, source, "$", selectTree), nil
}

func buildSelectFields(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	switch field.FieldType {
	case StringFieldType, NumberFieldType, BooleanFieldType, DateTimeFieldType:
		// If root of source (path is just $) is primitive just return the source
		if path == "$" {
			return source
		}
		return sqlVariant.JSONExtract(source, path)
	case CollectionFieldType:
		if field.CollectionItemMeta.FieldType == RelationshipFieldType {
			// This is an optimisation to allow us to do a single
			// aggregate query to the foreign table instead of a
			// sub query per item in the collection.
			return buildSelectFieldsForRelationshipCollectionFieldType(sqlVariant, schemaMetas, field, identifier, source, path, st)
		}
		return buildSelectFieldsForCollectionFieldType(sqlVariant, schemaMetas, field, identifier, source, path, st)
	case ComplexFieldType:
		return buildSelectFieldsForComplexFieldType(sqlVariant, schemaMetas, field, identifier, source, path, st)
	case RelationshipFieldType:
		return buildSelectFieldsForRelationshipFieldType(sqlVariant, schemaMetas, field, identifier, source, path, st)
	default:
		log.Errorf("Unsupported field type %v", field.FieldType)
		// TODO(sambetts) Return an error here
		return ""
	}
}

func buildSelectFieldsForRelationshipCollectionFieldType(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	if st == nil || !st.expand {
		return sqlVariant.JSONExtract(source, path)
	}

	schemaName := field.CollectionItemMeta.RelationshipSchema
	schema := schemaMetas[schemaName]
	newSource := fmt.Sprintf("%s.Data", schema.Table)

	where := fmt.Sprintf(
		"WHERE %s = %s",
		sqlVariant.JSONExtract(newSource, fmt.Sprintf("$.%s", field.CollectionItemMeta.RelationshipProperty)),
		sqlVariant.JSONExtract(
			fmt.Sprintf("%s.value", identifier),
			fmt.Sprintf("$.%s", field.CollectionItemMeta.RelationshipProperty),
		),
	)
	if st != nil {
		if st.filter != nil {
			conditions, _ := buildWhereFromFilter(sqlVariant, schemaMetas, field, newSource, newSource, st.filter.Tree, "")
			where = fmt.Sprintf("%s and %s", where, conditions)
		}
	}

	parts := []string{}
	for key, fm := range schema.Fields {
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
		sel := st.children[key]

		extract := buildSelectFields(sqlVariant, schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), newSource, fmt.Sprintf("$.%s", key), sel)
		part := fmt.Sprintf("'%s', %s", key, sqlVariant.JSONCast(extract))
		parts = append(parts, part)
	}
	subQuery := sqlVariant.JSONObject(parts)

	return fmt.Sprintf("(SELECT %s FROM %s,%s AS %s %s)", sqlVariant.JSONArrayAggregate(subQuery), schema.Table, sqlVariant.JSONEach(sqlVariant.JSONExtract(source, path)), identifier, where)
}

func buildSelectFieldsForRelationshipFieldType(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	if st == nil || !st.expand {
		return sqlVariant.JSONExtract(source, path)
	}

	schemaName := field.RelationshipSchema
	schema := schemaMetas[schemaName]
	newsource := fmt.Sprintf("%s.Data", schema.Table)
	parts := []string{}
	for key, fm := range schema.Fields {
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
		sel := st.children[key]

		extract := buildSelectFields(sqlVariant, schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), newsource, fmt.Sprintf("$.%s", key), sel)
		part := fmt.Sprintf("'%s', %s", key, sqlVariant.JSONCast(extract))
		parts = append(parts, part)
	}
	object := sqlVariant.JSONObject(parts)

	return fmt.Sprintf("(SELECT %s FROM %s WHERE %s = %s)", object, schema.Table,
		sqlVariant.JSONExtract(newsource, fmt.Sprintf("$.%s", field.RelationshipProperty)),
		sqlVariant.JSONExtract(source, fmt.Sprintf("%s.%s", path, field.RelationshipProperty)),
	)
}

func getDiscriminatorValue(schemaName string, field FieldMeta) string {
	if t, ok := field.DiscriminatorSchemaMapping[schemaName]; ok {
		return t
	}
	return schemaName
}

// nolint:cyclop
func buildSelectFieldsForComplexFieldType(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	// If there are no children in the select tree for this complex
	// type, shortcircuit and just return the data from the DB raw,
	// as there is no need to build the complex query, and it'll
	// ensure that null values are handled correctly.
	if st == nil || len(st.children) == 0 {
		return sqlVariant.JSONExtract(source, path)
	}

	objects := []string{}
	for _, schemaName := range field.ComplexFieldSchemas {
		schema := schemaMetas[schemaName]

		parts := []string{}
		if field.DiscriminatorProperty != "" {
			objectType := getDiscriminatorValue(schemaName, field)
			parts = append(parts, fmt.Sprintf("'%s', '%s'", field.DiscriminatorProperty, objectType))
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

			extract := buildSelectFields(sqlVariant, schemaMetas, fm, fmt.Sprintf("%s%s", identifier, key), source, fmt.Sprintf("%s.%s", path, key), sel)
			part := fmt.Sprintf("'%s', %s", key, sqlVariant.JSONCast(extract))
			parts = append(parts, part)
		}
		objects = append(objects, sqlVariant.JSONObject(parts))
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
		"(SELECT %s.value FROM %s AS %s WHERE %s = %s)",
		identifier, sqlVariant.JSONEach(sqlVariant.JSONArray(objects)), identifier,
		sqlVariant.JSONExtract(fmt.Sprintf("%s.value", identifier), fmt.Sprintf("$.%s", field.DiscriminatorProperty)),
		sqlVariant.JSONExtract(source, fmt.Sprintf("%s.%s", path, field.DiscriminatorProperty)),
	)
}

func buildSelectFieldsForCollectionFieldType(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier, source, path string, st *selectNode) string {
	newIdentifier := fmt.Sprintf("%sOptions", identifier)
	newSource := fmt.Sprintf("%s.value", newIdentifier)

	var where string
	var orderby string
	var newSelectNode *selectNode
	if st != nil {
		if st.filter != nil {
			conditions, _ := buildWhereFromFilter(sqlVariant, schemaMetas, *field.CollectionItemMeta, fmt.Sprintf("%sFilter", identifier), newSource, st.filter.Tree, "")
			where = fmt.Sprintf("WHERE %s", conditions)
		}

		if st.orderby != nil {
			conditions, err := buildOrderByFromOdata(sqlVariant, schemaMetas, *field.CollectionItemMeta, fmt.Sprintf("%sFilter", identifier), newSource, st.orderby.OrderByItems)
			// TODO(sambetts) Add error handling to buildSelectFields
			if err != nil {
				log.Errorf("Failed to build DB query from $orderby: %v", err)
			}

			orderby = fmt.Sprintf("ORDER BY %s", conditions)
		}

		// OrderBy/Filter is handled on the outer collection,
		// but select/expand are handled when building the
		// subQuery for each item in the collection so we have
		// to pass that down.
		newSelectNode = st.clone()
		newSelectNode.filter = nil
		newSelectNode.orderby = nil
	}

	subQuery := buildSelectFields(sqlVariant, schemaMetas, *field.CollectionItemMeta, fmt.Sprintf("%sOptions", newIdentifier), newSource, "$", newSelectNode)

	// This query will produce an exploded list of items (one row per item) from the collection, selected, filtered and ordered
	listQuery := fmt.Sprintf("SELECT %s AS value FROM %s AS %s %s %s", subQuery, sqlVariant.JSONEach(sqlVariant.JSONExtract(source, path)), newIdentifier, where, orderby)

	// Now aggregate all the rows back into a JSON array
	var aggregateValue string
	switch field.CollectionItemMeta.FieldType {
	case StringFieldType, NumberFieldType, BooleanFieldType, DateTimeFieldType:
		aggregateValue = fmt.Sprintf("%s.value", identifier)
	case CollectionFieldType, RelationshipFieldType, ComplexFieldType:
		fallthrough
	default:
		// For non-primitives use -> '$' to convert the value back to a
		// json object in the aggregate.
		aggregateValue = sqlVariant.JSONExtract(fmt.Sprintf("%s.value", identifier), "$")
	}
	return fmt.Sprintf("(SELECT %s FROM (%s) AS %s)", sqlVariant.JSONArrayAggregate(aggregateValue), listQuery, identifier)
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
		return "", fmt.Errorf("unsupported token type %s", node.Token.Type)
	}
}

// nolint:cyclop
func expandItemsToReachPath(schemaMetas map[string]SchemaMeta, field FieldMeta, currentPath, path string) string {
	switch field.FieldType {
	case StringFieldType, NumberFieldType, BooleanFieldType, DateTimeFieldType:
		return ""
	case CollectionFieldType:
		return expandItemsToReachPath(schemaMetas, *field.CollectionItemMeta, currentPath, path)
	case RelationshipFieldType:
		schema := schemaMetas[field.RelationshipSchema]
		fieldName, pathRemainder, _ := strings.Cut(path, "/")
		newfield := schema.Fields[fieldName]

		newPath := fieldName
		if currentPath != "" {
			newPath = fmt.Sprintf("%s/%s", currentPath, fieldName)
		}

		otherExpands := expandItemsToReachPath(schemaMetas, newfield, newPath, pathRemainder)

		expands := currentPath
		if otherExpands != "" {
			expands = fmt.Sprintf("%s,%s", currentPath, otherExpands)
		}
		return expands
	case ComplexFieldType:
		// We've reached the bottom of the path and its a complex type
		// so isn't in expanded.
		if path == "" {
			return ""
		}

		expands := []string{}
		fieldName, pathRemainder, _ := strings.Cut(path, "/")
		for _, schemaName := range field.ComplexFieldSchemas {
			schema := schemaMetas[schemaName]

			newField, ok := schema.Fields[fieldName]
			if !ok {
				continue
			}

			newPath := fieldName
			if currentPath != "" {
				newPath = fmt.Sprintf("%s/%s", currentPath, fieldName)
			}

			otherExpands := expandItemsToReachPath(schemaMetas, newField, newPath, pathRemainder)

			if otherExpands != "" {
				expands = append(expands, otherExpands)
			}
		}
		return strings.Join(expands, ",")
	default:
		return ""
	}
}

func buildWhereFromFilter(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, node *godata.ParseNode, chompPathPrefix string) (string, error) {
	tokenType := node.Token.Type

	switch tokenType {
	case godata.ExpressionTokenLogical:
		return buildWhereFromLogicalOperator(sqlVariant, schemaMetas, field, identifier, source, node, chompPathPrefix)
	case godata.ExpressionTokenFunc:
		return buildWhereFromFunction(sqlVariant, schemaMetas, field, identifier, source, node, chompPathPrefix)
	case godata.ExpressionTokenLambdaNav:
		return buildWhereFromLambda(sqlVariant, schemaMetas, field, identifier, source, node, chompPathPrefix)
	case godata.ExpressionTokenString, godata.ExpressionTokenInteger, godata.ExpressionTokenFloat, godata.ExpressionTokenBoolean, godata.ExpressionTokenDateTime, godata.ExpressionTokenNull:
		return buildWhereFromPrimative(sqlVariant, node)
	case godata.ExpressionTokenNav, godata.ExpressionTokenLiteral:
		return buildWhereFromNav(sqlVariant, schemaMetas, field, identifier, source, node, chompPathPrefix)
	default:
		return "", fmt.Errorf("unexpected Token Type: %s", tokenType)
	}
}

func buildWhereFromPrimative(sqlVariant jsonsql.Variant, node *godata.ParseNode) (string, error) {
	tokenType := node.Token.Type
	switch tokenType {
	case godata.ExpressionTokenString:
		return sqlVariant.JSONQuote(node.Token.Value), nil
	case godata.ExpressionTokenInteger, godata.ExpressionTokenFloat:
		return sqlVariant.CastToInteger(node.Token.Value), nil
	case godata.ExpressionTokenBoolean:
		return singleQuote(node.Token.Value), nil
	case godata.ExpressionTokenDateTime:
		return sqlVariant.CastToDateTime(singleQuote(node.Token.Value)), nil
	case godata.ExpressionTokenNull:
		return "NULL", nil
	default:
		return "", fmt.Errorf("unexpected token type: %s", tokenType)
	}
}

func buildWhereFromNav(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, node *godata.ParseNode, chompPathPrefix string) (string, error) {
	// Convert ODATA paths with slashes like "Thing/Name" into JSON
	// path like "Thing.Name".
	queryPath, err := buildJSONPathFromParseNode(node)
	if err != nil {
		return "", fmt.Errorf("unable to covert oData path to json path: %w", err)
	}
	queryPath = strings.TrimPrefix(queryPath, chompPathPrefix)

	fieldSource, err := sourceFromQueryPath(sqlVariant, schemaMetas, field, identifier, source, queryPath)
	if err != nil {
		return "", fmt.Errorf("unable to build source for filter %w", err)
	}

	fieldMetas, err := fieldMetaFromQueryPath(schemaMetas, field, queryPath)
	if err != nil {
		return "", fmt.Errorf("error finding field meta in schema for query path %s: %w", queryPath, err)
	}
	if len(fieldMetas) < 1 {
		return "", fmt.Errorf("unable to find field meta in schema for query path %s", queryPath)
	}
	fieldMeta := fieldMetas[0]

	switch fieldMeta.FieldType {
	case StringFieldType, BooleanFieldType:
		return sqlVariant.JSONExtract(fieldSource, queryPath), nil
	case NumberFieldType:
		return sqlVariant.CastToInteger(sqlVariant.JSONExtractText(fieldSource, queryPath)), nil
	case DateTimeFieldType:
		return sqlVariant.CastToDateTime(sqlVariant.JSONExtractText(fieldSource, queryPath)), nil
	case CollectionFieldType, RelationshipFieldType, ComplexFieldType:
		return sqlVariant.JSONExtractText(fieldSource, queryPath), nil
	default:
		return "", fmt.Errorf("unable to directly extract field type %s, you might need a function like length() or to specify a specific field", fieldMeta.FieldType)
	}
}

// nolint:cyclop
func buildWhereFromLogicalOperator(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, node *godata.ParseNode, chompPathPrefix string) (string, error) {
	operator := node.Token.Value
	var query string
	switch operator {
	case "eq", "ne", "gt", "ge", "lt", "le":
		left, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[0], chompPathPrefix)
		if err != nil {
			return query, err
		}

		right, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[1], chompPathPrefix)
		if err != nil {
			return query, err
		}

		sqlOperator := sqlOperators[operator]
		switch node.Children[1].Token.Type {
		case godata.ExpressionTokenNull:
			if operator == "eq" {
				sqlOperator = "is"
			} else if operator == "ne" {
				sqlOperator = "is not"
			} else {
				return "", fmt.Errorf("unsupported ExpressionTokenNull operator %s", operator)
			}
		}

		query = fmt.Sprintf("(%s %s %s)", left, sqlOperator, right)
	case "and":
		left, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[0], chompPathPrefix)
		if err != nil {
			return query, err
		}
		right, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[1], chompPathPrefix)
		if err != nil {
			return query, err
		}
		query = fmt.Sprintf("(%s AND %s)", left, right)
	case "or":
		left, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[0], chompPathPrefix)
		if err != nil {
			return query, err
		}
		right, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[1], chompPathPrefix)
		if err != nil {
			return query, err
		}
		query = fmt.Sprintf("(%s OR %s)", left, right)
	case "not":
		subquery, err := buildWhereFromFilter(sqlVariant, schemaMetas, field, identifier, source, node.Children[0], chompPathPrefix)
		if err != nil {
			return query, err
		}
		query = fmt.Sprintf("NOT (%s)", subquery)
	default:
		return query, fmt.Errorf("unsupported operator: %s", operator)
	}

	return query, nil
}

func buildWhereFromFunction(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, node *godata.ParseNode, chompPathPrefix string) (string, error) {
	function := node.Token.Value
	var query string
	switch function {
	case "contains", "endswith", "startswith":
		// Convert ODATA paths with slashes like "Thing/Name" into JSON
		// path like "Thing.Name".
		queryPath, err := buildJSONPathFromParseNode(node.Children[0])
		if err != nil {
			return "", fmt.Errorf("unable to covert oData path to json path: %w", err)
		}
		queryPath = strings.TrimPrefix(queryPath, chompPathPrefix)

		fieldSource, err := sourceFromQueryPath(sqlVariant, schemaMetas, field, identifier, source, queryPath)
		if err != nil {
			return "", fmt.Errorf("unable to build source for filter %w", err)
		}

		right := node.Children[1].Token.Value
		var value interface{}
		switch node.Children[1].Token.Type {
		case godata.ExpressionTokenString:
			r := strings.ReplaceAll(right, "'", "")
			value = fmt.Sprintf(sqlOperators[function], r)
		default:
			return query, fmt.Errorf("unsupported token type %s", node.Children[1].Token.Type)
		}

		query = fmt.Sprintf(
			"%s LIKE '%s'",
			sqlVariant.JSONExtractText(fieldSource, fmt.Sprintf("$.%s", queryPath)),
			value,
		)
	case "length":
		// Convert ODATA paths with slashes like "Thing/Name" into JSON
		// path like "Thing.Name".
		queryPath, err := buildJSONPathFromParseNode(node.Children[0])
		if err != nil {
			return "", fmt.Errorf("unable to covert oData path to json path: %w", err)
		}
		queryPath = strings.TrimPrefix(queryPath, chompPathPrefix)

		fieldSource, err := sourceFromQueryPath(sqlVariant, schemaMetas, field, identifier, source, queryPath)
		if err != nil {
			return "", fmt.Errorf("unable to build source for filter %w", err)
		}

		return sqlVariant.JSONArrayLength(sqlVariant.JSONExtract(fieldSource, queryPath)), nil
	default:
		return query, fmt.Errorf("unsupported function: %s", function)
	}

	return query, nil
}

// nolint:cyclop
func fieldMetaFromQueryPath(schemaMetas map[string]SchemaMeta, field FieldMeta, path string) ([]FieldMeta, error) {
	switch field.FieldType {
	case StringFieldType, NumberFieldType, BooleanFieldType, DateTimeFieldType:
		if path == "" {
			return []FieldMeta{field}, nil
		} else {
			return nil, fmt.Errorf("can not subpath a primitive type")
		}
	case CollectionFieldType:
		return fieldMetaFromQueryPath(schemaMetas, *field.CollectionItemMeta, path)
	case RelationshipFieldType:
		schema := schemaMetas[field.RelationshipSchema]
		fieldName, pathRemainder, _ := strings.Cut(path, ".")
		newfield := schema.Fields[fieldName]
		return fieldMetaFromQueryPath(schemaMetas, newfield, pathRemainder)
	case ComplexFieldType:
		if path == "" {
			return []FieldMeta{field}, nil
		}

		fieldMetas := []FieldMeta{}
		fieldName, pathRemainder, _ := strings.Cut(path, ".")
		for _, schemaName := range field.ComplexFieldSchemas {
			schema := schemaMetas[schemaName]
			newField, ok := schema.Fields[fieldName]
			if !ok {
				continue
			}
			otherMetas, err := fieldMetaFromQueryPath(schemaMetas, newField, pathRemainder)
			if err != nil {
				return nil, fmt.Errorf("can not get field meta from path for schema %s:%s %s", schemaName, fieldName, pathRemainder)
			}
			fieldMetas = append(fieldMetas, otherMetas...)
		}
		return fieldMetas, nil
	default:
		return nil, fmt.Errorf("unknown field type: %v", field.FieldType)
	}
}

func buildWhereFromLambda(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, node *godata.ParseNode, chompPathPrefix string) (string, error) {
	sourcePath, err := buildJSONPathFromParseNode(node.Children[0])
	if err != nil {
		return "", fmt.Errorf("unable to covert oData path to json path: %w", err)
	}
	sourcePath = strings.TrimPrefix(sourcePath, chompPathPrefix)

	fieldSource, err := sourceFromQueryPath(sqlVariant, schemaMetas, field, identifier, source, sourcePath)
	if err != nil {
		return "", fmt.Errorf("unable to build source for filter %w", err)
	}

	newFieldMetas, err := fieldMetaFromQueryPath(schemaMetas, field, sourcePath)
	if err != nil {
		return "", fmt.Errorf("error finding field meta in schema for query path %s: %w", sourcePath, err)
	}
	if len(newFieldMetas) < 1 {
		return "", fmt.Errorf("unable to find field meta in schema for query path %s", sourcePath)
	}

	lambda := node.Children[1]
	lambdaFunc := lambda.Token.Value
	switch lambdaFunc {
	case "any":
		newIdentifier := fmt.Sprintf("%sFilterOptions", identifier)
		subquery, err := buildWhereFromFilter(sqlVariant, schemaMetas, newFieldMetas[0], newIdentifier, fmt.Sprintf("%s.value", identifier), lambda.Children[1], fmt.Sprintf("%s.", lambda.Children[0].Token.Value))
		if err != nil {
			return "", fmt.Errorf("unable to build query inside lambda: %w", err)
		}
		return fmt.Sprintf("EXISTS (SELECT 1 FROM %s AS %s WHERE %s)", sqlVariant.JSONEach(sqlVariant.JSONExtract(fieldSource, sourcePath)), identifier, subquery), nil
	case "all":
		newIdentifier := fmt.Sprintf("%sFilterOptions", identifier)
		subquery, err := buildWhereFromFilter(sqlVariant, schemaMetas, newFieldMetas[0], newIdentifier, fmt.Sprintf("%s.value", identifier), lambda.Children[1], fmt.Sprintf("%s.", lambda.Children[0].Token.Value))
		if err != nil {
			return "", fmt.Errorf("unable to build query inside lambda: %w", err)
		}
		return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s AS %s WHERE NOT (%s))", sqlVariant.JSONEach(sqlVariant.JSONExtract(fieldSource, sourcePath)), identifier, subquery), nil
	default:
		return "", fmt.Errorf("unknown lambda function %v", lambdaFunc)
	}
}

func sourceFromQueryPath(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, queryPath string) (string, error) {
	log.Debugf("QueryPath: %s", queryPath)

	// ODATA path that would be present if we were to $select the
	// field being filtered
	selectPath := strings.ReplaceAll(queryPath, ".", "/")

	// Calculate any $expands that would be required to reach that
	// selected path
	expandItems := expandItemsToReachPath(schemaMetas, field, "", selectPath)

	// If this filter requires expanded items then build a
	// JSON_OBJECT with the just the required fields selected and
	// expanded to perform the filter against.
	fieldSource := source
	if expandItems != "" {
		var err error
		fieldSource, err = buildSelectFieldsFromSelectAndExpand(sqlVariant, schemaMetas, field, identifier, source, &selectPath, &expandItems)
		if err != nil {
			return "", fmt.Errorf("unable to build source %w", err)
		}
	}
	return fieldSource, nil
}

func buildOrderByFromOdata(sqlVariant jsonsql.Variant, schemaMetas map[string]SchemaMeta, field FieldMeta, identifier string, source string, orderbyItems []*godata.OrderByItem) (string, error) {
	conditions := []string{}

	for _, item := range orderbyItems {
		queryPath, err := buildJSONPathFromParseNode(item.Tree.Tree)
		if err != nil {
			return "", fmt.Errorf("failed to convert odata path to json path: %w", err)
		}

		fieldSource, err := sourceFromQueryPath(sqlVariant, schemaMetas, field, identifier, source, queryPath)
		if err != nil {
			return "", fmt.Errorf("unable to build source for filter %w", err)
		}

		conditions = append(conditions, fmt.Sprintf(
			"%s %s",
			sqlVariant.JSONExtractText(fieldSource, fmt.Sprintf("$.%s", queryPath)),
			strings.ToUpper(item.Order)),
		)
	}

	return strings.Join(conditions, ", "), nil
}
