package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestServeStdioHandlesInitializeToolsListAndCall(t *testing.T) {
	s := NewServer(ToolContext{})
	if err := s.RegisterTool(ToolDefinition{
		Name:        "plan_ai.project_status",
		Description: "Get project status",
		Schema:      JSONSchema{Type: "object"},
		Handler: func(args map[string]any) (*ToolResult, error) {
			return &ToolResult{Success: true, Data: map[string]any{"status": "initialized"}}, nil
		},
	}); err != nil {
		t.Fatalf("register tool: %v", err)
	}

	input := encodeMCPMessage(t, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}}) +
		encodeMCPMessage(t, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}) +
		encodeMCPMessage(t, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{"name": "plan_ai.project_status", "arguments": map[string]any{}}})

	var out bytes.Buffer
	if err := ServeStdio(strings.NewReader(input), &out, s); err != nil {
		t.Fatalf("ServeStdio: %v", err)
	}

	responses := decodeMCPMessages(t, out.String())
	if len(responses) != 3 {
		t.Fatalf("responses len = %d, want 3; raw=%s", len(responses), out.String())
	}
	assertJSONRPCID(t, responses[0], float64(1))
	assertJSONRPCID(t, responses[1], float64(2))
	assertJSONRPCID(t, responses[2], float64(3))

	result, ok := responses[1]["result"].(map[string]any)
	if !ok {
		t.Fatalf("tools/list response missing result: %#v", responses[1])
	}
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools/list tools = %#v", result["tools"])
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "plan_ai.project_status" {
		t.Fatalf("tool name = %v", tool["name"])
	}

	callResult := responses[2]["result"].(map[string]any)
	content := callResult["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "initialized") {
		t.Fatalf("tools/call text missing initialized: %s", text)
	}
}

func encodeMCPMessage(t *testing.T, msg map[string]any) string {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
}

func decodeMCPMessages(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var responses []map[string]any
	for raw != "" {
		sep := strings.Index(raw, "\r\n\r\n")
		if sep < 0 {
			t.Fatalf("missing header separator in %q", raw)
		}
		header := raw[:sep]
		var length int
		if _, err := fmt.Sscanf(header, "Content-Length: %d", &length); err != nil {
			t.Fatalf("parse header %q: %v", header, err)
		}
		start := sep + len("\r\n\r\n")
		end := start + length
		if end > len(raw) {
			t.Fatalf("message length %d exceeds raw length", length)
		}
		var msg map[string]any
		if err := json.Unmarshal([]byte(raw[start:end]), &msg); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		responses = append(responses, msg)
		raw = raw[end:]
	}
	return responses
}

func assertJSONRPCID(t *testing.T, msg map[string]any, want float64) {
	t.Helper()
	if msg["jsonrpc"] != "2.0" || msg["id"] != want {
		t.Fatalf("response id/jsonrpc = %#v", msg)
	}
}
