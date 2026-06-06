package main

import (
	"strings"
	"testing"
)

func TestRemoteUpdateScriptInstallsSinglePlanAIBinary(t *testing.T) {
	script := remoteUpdateScript()
	if strings.Contains(script, "plan-ai-mcp-server") {
		t.Fatalf("remote update script should not install legacy MCP binary:\n%s", script)
	}
	if !strings.Contains(script, `go build -o "${BIN_DIR}/plan-ai" ./cmd/plan-ai`) {
		t.Fatalf("remote update script should build plan-ai CLI:\n%s", script)
	}
}
