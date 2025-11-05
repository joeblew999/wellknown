package pbmcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pocketbase/pocketbase/core"
)

// registerTools registers all MCP tools for PocketBase operations
func (s *Server) registerTools() {
	// Tool: List all collections
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_collections",
		Description: "List all PocketBase collections with their schemas",
	}, s.handleListCollections)

	// Tool: Query records from a collection
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "query_records",
		Description: "Query records from a PocketBase collection with optional filtering, sorting, and pagination",
	}, s.handleQueryRecords)

	// Tool: Get a single record by ID
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_record",
		Description: "Get a specific record by its ID from a collection",
	}, s.handleGetRecord)

	// Tool: Create a new record
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "create_record",
		Description: "Create a new record in a PocketBase collection",
	}, s.handleCreateRecord)

	// Tool: Update a record
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "update_record",
		Description: "Update an existing record in a PocketBase collection",
	}, s.handleUpdateRecord)

	// Tool: Delete a record
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "delete_record",
		Description: "Delete a record from a PocketBase collection",
	}, s.handleDeleteRecord)
}

// Input/Output types for tools

type ListCollectionsInput struct{}

type ListCollectionsOutput struct {
	Collections []CollectionInfo `json:"collections" jsonschema:"List of all collections"`
}

type CollectionInfo struct {
	ID     string   `json:"id" jsonschema:"Collection ID"`
	Name   string   `json:"name" jsonschema:"Collection name"`
	Type   string   `json:"type" jsonschema:"Collection type (base, auth, view)"`
	Fields []string `json:"fields" jsonschema:"List of field names"`
}

type QueryRecordsInput struct {
	Collection string `json:"collection" jsonschema:"required,The collection name to query"`
	Filter     string `json:"filter,omitempty" jsonschema:"Optional PocketBase filter query (e.g., 'status = \"active\"')"`
	Sort       string `json:"sort,omitempty" jsonschema:"Optional sort order (e.g., '-created' for descending)"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of records to return (default 50)"`
	Page       int    `json:"page,omitempty" jsonschema:"Page number for pagination (default 1)"`
}

type QueryRecordsOutput struct {
	Records    []map[string]interface{} `json:"records" jsonschema:"The matching records"`
	TotalItems int                      `json:"totalItems" jsonschema:"Total number of matching records"`
	Page       int                      `json:"page" jsonschema:"Current page number"`
	PerPage    int                      `json:"perPage" jsonschema:"Records per page"`
	TotalPages int                      `json:"totalPages" jsonschema:"Total number of pages"`
}

type GetRecordInput struct {
	Collection string `json:"collection" jsonschema:"required,The collection name"`
	RecordID   string `json:"record_id" jsonschema:"required,The record ID"`
}

type GetRecordOutput struct {
	Record map[string]interface{} `json:"record" jsonschema:"The record data"`
}

type CreateRecordInput struct {
	Collection string                 `json:"collection" jsonschema:"required,The collection name"`
	Data       map[string]interface{} `json:"data" jsonschema:"required,The record data to create"`
}

type CreateRecordOutput struct {
	Record map[string]interface{} `json:"record" jsonschema:"The created record"`
}

type UpdateRecordInput struct {
	Collection string                 `json:"collection" jsonschema:"required,The collection name"`
	RecordID   string                 `json:"record_id" jsonschema:"required,The record ID to update"`
	Data       map[string]interface{} `json:"data" jsonschema:"required,The data to update"`
}

type UpdateRecordOutput struct {
	Record map[string]interface{} `json:"record" jsonschema:"The updated record"`
}

type DeleteRecordInput struct {
	Collection string `json:"collection" jsonschema:"required,The collection name"`
	RecordID   string `json:"record_id" jsonschema:"required,The record ID to delete"`
}

type DeleteRecordOutput struct {
	Success bool   `json:"success" jsonschema:"Whether the deletion was successful"`
	Message string `json:"message" jsonschema:"Status message"`
}

// Tool handlers

func (s *Server) handleListCollections(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListCollectionsInput,
) (*mcp.CallToolResult, ListCollectionsOutput, error) {
	collections, err := s.app.FindAllCollections()
	if err != nil {
		return nil, ListCollectionsOutput{}, fmt.Errorf("failed to list collections: %w", err)
	}

	var result []CollectionInfo
	for _, col := range collections {
		fields := []string{}
		for _, field := range col.Fields {
			fields = append(fields, field.GetName())
		}

		result = append(result, CollectionInfo{
			ID:     col.Id,
			Name:   col.Name,
			Type:   col.Type,
			Fields: fields,
		})
	}

	return nil, ListCollectionsOutput{Collections: result}, nil
}

func (s *Server) handleQueryRecords(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input QueryRecordsInput,
) (*mcp.CallToolResult, QueryRecordsOutput, error) {
	// Set defaults
	if input.Limit == 0 {
		input.Limit = 50
	}
	if input.Page == 0 {
		input.Page = 1
	}

	// Query records with pagination
	records, err := s.app.FindRecordsByFilter(
		input.Collection,
		input.Filter,
		input.Sort,
		input.Limit,
		(input.Page-1)*input.Limit,
	)
	if err != nil {
		return nil, QueryRecordsOutput{}, fmt.Errorf("failed to query records: %w", err)
	}

	// Convert to map format
	var result []map[string]interface{}
	for _, record := range records {
		result = append(result, record.PublicExport())
	}

	// Get total count for pagination
	totalItems := len(result) // Simplified - ideally should count total matching records
	totalPages := (totalItems + input.Limit - 1) / input.Limit

	return nil, QueryRecordsOutput{
		Records:    result,
		TotalItems: totalItems,
		Page:       input.Page,
		PerPage:    input.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *Server) handleGetRecord(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetRecordInput,
) (*mcp.CallToolResult, GetRecordOutput, error) {
	record, err := s.app.FindRecordById(input.Collection, input.RecordID)
	if err != nil {
		return nil, GetRecordOutput{}, fmt.Errorf("failed to get record: %w", err)
	}

	return nil, GetRecordOutput{Record: record.PublicExport()}, nil
}

func (s *Server) handleCreateRecord(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreateRecordInput,
) (*mcp.CallToolResult, CreateRecordOutput, error) {
	collection, err := s.app.FindCollectionByNameOrId(input.Collection)
	if err != nil {
		return nil, CreateRecordOutput{}, fmt.Errorf("collection not found: %w", err)
	}

	record := core.NewRecord(collection)

	// Set fields from input data
	for key, value := range input.Data {
		record.Set(key, value)
	}

	// Save the record
	if err := s.app.Save(record); err != nil {
		return nil, CreateRecordOutput{}, fmt.Errorf("failed to create record: %w", err)
	}

	return nil, CreateRecordOutput{Record: record.PublicExport()}, nil
}

func (s *Server) handleUpdateRecord(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateRecordInput,
) (*mcp.CallToolResult, UpdateRecordOutput, error) {
	record, err := s.app.FindRecordById(input.Collection, input.RecordID)
	if err != nil {
		return nil, UpdateRecordOutput{}, fmt.Errorf("record not found: %w", err)
	}

	// Update fields from input data
	for key, value := range input.Data {
		record.Set(key, value)
	}

	// Save the updated record
	if err := s.app.Save(record); err != nil {
		return nil, UpdateRecordOutput{}, fmt.Errorf("failed to update record: %w", err)
	}

	return nil, UpdateRecordOutput{Record: record.PublicExport()}, nil
}

func (s *Server) handleDeleteRecord(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DeleteRecordInput,
) (*mcp.CallToolResult, DeleteRecordOutput, error) {
	record, err := s.app.FindRecordById(input.Collection, input.RecordID)
	if err != nil {
		return nil, DeleteRecordOutput{}, fmt.Errorf("record not found: %w", err)
	}

	if err := s.app.Delete(record); err != nil {
		return nil, DeleteRecordOutput{}, fmt.Errorf("failed to delete record: %w", err)
	}

	return nil, DeleteRecordOutput{
		Success: true,
		Message: fmt.Sprintf("Record %s deleted successfully", input.RecordID),
	}, nil
}
