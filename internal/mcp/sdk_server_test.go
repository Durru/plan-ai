package mcp

import (
	"testing"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func TestNewSDKServerReturnsMark3LabsServer(t *testing.T) {
	srv, err := NewSDKServer(ToolContext{}, DefaultToolDependencies(), false)
	if err != nil {
		t.Fatalf("NewSDKServer: %v", err)
	}
	if _, ok := any(srv).(*mcpserver.MCPServer); !ok {
		t.Fatalf("NewSDKServer returned %T, want *server.MCPServer", srv)
	}
	if srv.GetTool("plan_ai.project_status") == nil {
		t.Fatal("SDK server missing plan_ai.project_status")
	}
}
