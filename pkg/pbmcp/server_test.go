package pbmcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/joeblew999/wellknown/pkg/pbmcp/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var testImpl = &mcp.Implementation{
	Name:    "test-client",
	Version: "1.0.0",
}

// TestServerCreation tests that the server can be created successfully
func TestServerCreation(t *testing.T) {
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	server := NewServer(app)
	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.server == nil {
		t.Fatal("Expected MCP server to be initialized")
	}
}

// TestToolRegistration tests that all tools are registered correctly
func TestToolRegistration(t *testing.T) {
	ctx := context.Background()
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	// Create server and client with in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer(app)
	client := mcp.NewClient(testImpl, nil)

	// Connect both sides
	_, err = server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientSession.Close()

	// List tools
	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	// Verify all expected tools are registered
	expectedTools := map[string]bool{
		"list_collections": false,
		"query_records":    false,
		"get_record":       false,
		"create_record":    false,
		"update_record":    false,
		"delete_record":    false,
	}

	for _, tool := range result.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %s to be registered", name)
		}
	}
}

// TestResourceRegistration tests that all resources are registered correctly
func TestResourceRegistration(t *testing.T) {
	ctx := context.Background()
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	// Create server and client with in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer(app)
	client := mcp.NewClient(testImpl, nil)

	// Connect both sides
	_, err = server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientSession.Close()

	// List resources
	result, err := clientSession.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	// Verify expected resources
	expectedResources := map[string]bool{
		"pocketbase://schema/collections": false,
		"pocketbase://collections":        false,
	}

	for _, resource := range result.Resources {
		if _, ok := expectedResources[resource.URI]; ok {
			expectedResources[resource.URI] = true
		}
	}

	for uri, found := range expectedResources {
		if !found {
			t.Errorf("Expected resource %s to be registered", uri)
		}
	}
}

// TestListCollectionsTool tests the list_collections tool
func TestListCollectionsTool(t *testing.T) {
	ctx := context.Background()
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	// Create server and client with in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer(app)
	client := mcp.NewClient(testImpl, nil)

	// Connect both sides
	_, err = server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientSession.Close()

	// Call list_collections tool
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_collections",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("Failed to call list_collections: %v", err)
	}

	// Verify we got content back
	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	// Extract text from content
	var textContent string
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	if textContent == "" {
		t.Fatal("Expected text content in response")
	}

	// Parse the JSON response
	var output ListCollectionsOutput
	if err := json.Unmarshal([]byte(textContent), &output); err != nil {
		t.Fatalf("Failed to parse collections: %v", err)
	}

	// Should have at least our test collections
	if len(output.Collections) < 2 {
		t.Errorf("Expected at least 2 collections, got %d", len(output.Collections))
	}

	// Verify test collections are present
	collectionNames := make(map[string]bool)
	for _, col := range output.Collections {
		collectionNames[col.Name] = true
	}

	if !collectionNames["test_users"] {
		t.Error("Expected test_users collection")
	}
	if !collectionNames["test_posts"] {
		t.Error("Expected test_posts collection")
	}
}

// TestQueryRecordsTool tests the query_records tool
func TestQueryRecordsTool(t *testing.T) {
	ctx := context.Background()
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	// Create test records
	users := testutil.GetTestUsers()
	for _, user := range users {
		if _, err := testutil.CreateTestRecord(app, "test_users", testutil.UserToMap(user)); err != nil {
			t.Fatalf("Failed to create test record: %v", err)
		}
	}

	// Create server and client with in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer(app)
	client := mcp.NewClient(testImpl, nil)

	// Connect both sides
	_, err = server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientSession.Close()

	// Call query_records tool
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "query_records",
		Arguments: map[string]any{
			"collection": "test_users",
		},
	})
	if err != nil {
		t.Fatalf("Failed to call query_records: %v", err)
	}

	// Verify we got content back
	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	// Extract text from content
	var textContent string
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	// Parse the JSON response
	var queryResult QueryRecordsOutput
	if err := json.Unmarshal([]byte(textContent), &queryResult); err != nil {
		t.Fatalf("Failed to parse query result: %v", err)
	}

	if queryResult.TotalItems != 3 {
		t.Errorf("Expected 3 records, got %d", queryResult.TotalItems)
	}

	if len(queryResult.Records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(queryResult.Records))
	}
}

// TestCreateRecordTool tests the create_record tool
func TestCreateRecordTool(t *testing.T) {
	ctx := context.Background()
	app, err := testutil.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testutil.CleanupTestApp(app)

	// Create server and client with in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := NewServer(app)
	client := mcp.NewClient(testImpl, nil)

	// Connect both sides
	_, err = server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect server: %v", err)
	}

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientSession.Close()

	// Call create_record tool
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_record",
		Arguments: map[string]any{
			"collection": "test_users",
			"data": map[string]any{
				"name":  "Test User",
				"email": "test@example.com",
				"age":   28,
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to call create_record: %v", err)
	}

	// Verify we got content back
	if len(result.Content) == 0 {
		t.Fatal("Expected content in response")
	}

	// Extract text from content
	var textContent string
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	// Parse the JSON response
	var output CreateRecordOutput
	if err := json.Unmarshal([]byte(textContent), &output); err != nil {
		t.Fatalf("Failed to parse created record: %v", err)
	}

	// Verify record has ID
	if output.Record["id"] == nil || output.Record["id"] == "" {
		t.Error("Expected record to have an ID")
	}

	// Verify data
	if output.Record["name"] != "Test User" {
		t.Errorf("Expected name to be 'Test User', got %v", output.Record["name"])
	}
}
