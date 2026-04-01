package mcp

import (
	"testing"
)

func TestTransportTypes(t *testing.T) {
	if TransportStdio != "stdio" {
		t.Errorf("TransportStdio = %v, want %v", TransportStdio, "stdio")
	}
	if TransportSSE != "sse" {
		t.Errorf("TransportSSE = %v, want %v", TransportSSE, "sse")
	}
	if TransportHTTP != "http" {
		t.Errorf("TransportHTTP = %v, want %v", TransportHTTP, "http")
	}
	if TransportWS != "ws" {
		t.Errorf("TransportWS = %v, want %v", TransportWS, "ws")
	}
}

func TestConfigScopes(t *testing.T) {
	if ConfigScopeLocal != "local" {
		t.Errorf("ConfigScopeLocal = %v, want %v", ConfigScopeLocal, "local")
	}
	if ConfigScopeUser != "user" {
		t.Errorf("ConfigScopeUser = %v, want %v", ConfigScopeUser, "user")
	}
	if ConfigScopeProject != "project" {
		t.Errorf("ConfigScopeProject = %v, want %v", ConfigScopeProject, "project")
	}
}

func TestJSONRPCHelpers(t *testing.T) {
	// Test CreateRequest
	req := CreateRequest("tools/list", nil, "123")
	if req.JsonRPC != "2.0" {
		t.Errorf("JsonRPC = %v, want %v", req.JsonRPC, "2.0")
	}
	if req.Method != "tools/list" {
		t.Errorf("Method = %v, want %v", req.Method, "tools/list")
	}
	if req.ID != "123" {
		t.Errorf("ID = %v, want %v", req.ID, "123")
	}

	// Test CreateResponse
	resp := CreateResponse("123", map[string]any{"tools": []any{}})
	if resp.JsonRPC != "2.0" {
		t.Errorf("JsonRPC = %v, want %v", resp.JsonRPC, "2.0")
	}
	if resp.ID != "123" {
		t.Errorf("ID = %v, want %v", resp.ID, "123")
	}
	if resp.Result == nil {
		t.Error("Result is nil")
	}

	// Test CreateError
	errResp := CreateError("123", -32601, "Method not found")
	if errResp.Error == nil {
		t.Fatal("Error is nil")
	}
	if errResp.Error.Code != -32601 {
		t.Errorf("Error.Code = %v, want %v", errResp.Error.Code, -32601)
	}

	// Test CreateNotification
	notify := CreateNotification("initialized", nil)
	if notify.ID != "" {
		t.Errorf("ID = %v, want empty for notification", notify.ID)
	}
	if notify.Method != "initialized" {
		t.Errorf("Method = %v, want %v", notify.Method, "initialized")
	}
}

func TestIsRequestNotificationResponse(t *testing.T) {
	req := &JSONRPCMessage{Method: "test", ID: "123"}
	if !IsRequest(req) {
		t.Error("IsRequest() = false, want true")
	}
	if IsNotification(req) {
		t.Error("IsNotification() = true, want false")
	}
	if IsResponse(req) {
		t.Error("IsResponse() = true, want false")
	}

	notify := &JSONRPCMessage{Method: "test"}
	if !IsNotification(notify) {
		t.Error("IsNotification() = false, want true")
	}

	resp := &JSONRPCMessage{Result: "ok", ID: "123"}
	if !IsResponse(resp) {
		t.Error("IsResponse() = false, want true")
	}
}

func TestParseTools(t *testing.T) {
	toolsData := []any{
		map[string]any{
			"name":        "test_tool",
			"description": "A test tool",
			"input_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"arg1": map[string]any{"type": "string"},
				},
			},
			"annotations": map[string]any{
				"readOnlyHint": true,
			},
		},
	}

	tools := parseTools(toolsData)
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(tools))
	}

	if tools[0].Name != "test_tool" {
		t.Errorf("tools[0].Name = %v, want %v", tools[0].Name, "test_tool")
	}
	if tools[0].Description != "A test tool" {
		t.Errorf("tools[0].Description = %v, want %v", tools[0].Description, "A test tool")
	}
	if tools[0].Annotations == nil || !tools[0].Annotations.ReadOnlyHint {
		t.Error("tools[0].Annotations.ReadOnlyHint = false, want true")
	}
}
