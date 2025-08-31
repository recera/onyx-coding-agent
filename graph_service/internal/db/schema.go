package db

import (
	"encoding/json"
	"fmt"

	"github.com/kuzudb/go-kuzu"
)

// PropertyInfo holds the name and data type of a property.
type PropertyInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// NodeSchema represents the schema of a node table.
type NodeSchema struct {
	Name       string         `json:"name"`
	Properties []PropertyInfo `json:"properties"`
}

// RelationshipSchema represents the schema of a relationship table.
type RelationshipSchema struct {
	Name        string         `json:"name"`
	Source      string         `json:"source"`
	Destination string         `json:"destination"`
	Properties  []PropertyInfo `json:"properties"`
}

// GraphSchema is the top-level container for the entire graph schema.
type GraphSchema struct {
	Nodes         []*NodeSchema         `json:"nodes"`
	Relationships []*RelationshipSchema `json:"relationships"`
}

// GetGraphSchema introspects the database and returns its complete schema.
func GetGraphSchema(conn *kuzu.Connection) (*GraphSchema, error) {
	schema := &GraphSchema{
		Nodes:         make([]*NodeSchema, 0),
		Relationships: make([]*RelationshipSchema, 0),
	}

	// 1. Get all tables (nodes and relationships)
	queryResult, err := conn.Query("CALL SHOW_TABLES()")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer queryResult.Close()

	for queryResult.HasNext() {
		row, err := queryResult.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next row: %w", err)
		}

		// Use GetValue(i) to access columns, not row[i]
		tableNameVal, err := row.GetValue(0)
		if err != nil {
			return nil, fmt.Errorf("failed to get tableName value: %w", err)
		}
		tableTypeVal, err := row.GetValue(1)
		if err != nil {
			return nil, fmt.Errorf("failed to get tableType value: %w", err)
		}

		tableName, _ := tableNameVal.(string)
		tableType, _ := tableTypeVal.(string)

		if tableType == "NODE" {
			nodeSchema, err := getNodeSchema(conn, tableName)
			if err != nil {
				return nil, err
			}
			schema.Nodes = append(schema.Nodes, nodeSchema)
		} else if tableType == "REL" {
			relSchema, err := getRelationshipSchema(conn, tableName)
			if err != nil {
				return nil, err
			}
			schema.Relationships = append(schema.Relationships, relSchema)
		}
	}

	return schema, nil
}

// getNodeSchema retrieves the detailed schema for a specific node table.
func getNodeSchema(conn *kuzu.Connection, tableName string) (*NodeSchema, error) {
	node := &NodeSchema{Name: tableName}

	queryResult, err := conn.Query(fmt.Sprintf("CALL TABLE_INFO('%s')", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get info for node table %s: %w", tableName, err)
	}
	defer queryResult.Close()

	for queryResult.HasNext() {
		row, err := queryResult.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next property row for %s: %w", tableName, err)
		}

		// Use GetValue(i) to access columns
		propNameVal, err := row.GetValue(1)
		if err != nil {
			return nil, fmt.Errorf("failed to get propName value for %s: %w", tableName, err)
		}
		propTypeVal, err := row.GetValue(2)
		if err != nil {
			return nil, fmt.Errorf("failed to get propType value for %s: %w", tableName, err)
		}

		propName, _ := propNameVal.(string)
		propType, _ := propTypeVal.(string)
		node.Properties = append(node.Properties, PropertyInfo{Name: propName, Type: propType})
	}
	return node, nil
}

// getRelationshipSchema retrieves the detailed schema for a specific relationship table.
func getRelationshipSchema(conn *kuzu.Connection, tableName string) (*RelationshipSchema, error) {
	rel := &RelationshipSchema{Name: tableName}

	// Get properties for the relationship table
	propsResult, err := conn.Query(fmt.Sprintf("CALL TABLE_INFO('%s')", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get info for rel table %s: %w", tableName, err)
	}

	for propsResult.HasNext() {
		row, err := propsResult.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next property row for %s: %w", tableName, err)
		}
		// Use GetValue(i) to access columns
		propNameVal, err := row.GetValue(1)
		if err != nil {
			return nil, fmt.Errorf("failed to get propName value for %s: %w", tableName, err)
		}
		propTypeVal, err := row.GetValue(2)
		if err != nil {
			return nil, fmt.Errorf("failed to get propType value for %s: %w", tableName, err)
		}

		propName, _ := propNameVal.(string)
		propType, _ := propTypeVal.(string)
		rel.Properties = append(rel.Properties, PropertyInfo{Name: propName, Type: propType})
	}
	propsResult.Close()

	// Get connections for the relationship table
	connResult, err := conn.Query(fmt.Sprintf("CALL SHOW_CONNECTION('%s')", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get connections for rel table %s: %w", tableName, err)
	}
	defer connResult.Close()

	if connResult.HasNext() {
		row, err := connResult.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next connection row for %s: %w", tableName, err)
		}

		// Use GetValue(i) to access columns
		sourceVal, err := row.GetValue(1)
		if err != nil {
			return nil, fmt.Errorf("failed to get source value for %s: %w", tableName, err)
		}
		destVal, err := row.GetValue(2)
		if err != nil {
			return nil, fmt.Errorf("failed to get destination value for %s: %w", tableName, err)
		}
		rel.Source, _ = sourceVal.(string)
		rel.Destination, _ = destVal.(string)
	}

	return rel, nil
}

// ToJsonString converts the GraphSchema to a nicely formatted JSON string.
func (gs *GraphSchema) ToJsonString() (string, error) {
	bytes, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}
	return string(bytes), nil
}
