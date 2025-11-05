package pbmcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerResources registers MCP resources (read-only data views)
func (s *Server) registerResources() {
	// Resource: Collections schema
	s.server.AddResource(
		&mcp.Resource{
			URI:         "pocketbase://schema/collections",
			Name:        "Collections Schema",
			Description: "All PocketBase collection definitions and schemas",
			MIMEType:    "application/json",
		},
		s.handleCollectionsSchemaResource,
	)

	// Resource: Collections list
	s.server.AddResource(
		&mcp.Resource{
			URI:         "pocketbase://collections",
			Name:        "Collections List",
			Description: "List of all collection names",
			MIMEType:    "application/json",
		},
		s.handleCollectionsListResource,
	)
}

func (s *Server) handleCollectionsSchemaResource(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	collections, err := s.app.FindAllCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to load collections: %w", err)
	}

	// Build schema info
	type FieldInfo struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Options map[string]interface{} `json:"options,omitempty"`
	}

	type CollectionSchema struct {
		ID     string      `json:"id"`
		Name   string      `json:"name"`
		Type   string      `json:"type"`
		Fields []FieldInfo `json:"fields"`
	}

	var schemas []CollectionSchema
	for _, col := range collections {
		var fields []FieldInfo
		for _, field := range col.Fields {
			fieldInfo := FieldInfo{
				Name: field.GetName(),
				Type: field.Type(),
			}
			fields = append(fields, fieldInfo)
		}

		schemas = append(schemas, CollectionSchema{
			ID:     col.Id,
			Name:   col.Name,
			Type:   col.Type,
			Fields: fields,
		})
	}

	data, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "pocketbase://schema/collections",
				MIMEType: "application/json",
				Text:     string(data),
			},
		},
	}, nil
}

func (s *Server) handleCollectionsListResource(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	collections, err := s.app.FindAllCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to load collections: %w", err)
	}

	var names []string
	for _, col := range collections {
		names = append(names, col.Name)
	}

	data, err := json.MarshalIndent(map[string]interface{}{
		"collections": names,
		"count":       len(names),
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal list: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "pocketbase://collections",
				MIMEType: "application/json",
				Text:     string(data),
			},
		},
	}, nil
}
