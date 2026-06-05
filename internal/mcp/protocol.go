package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ServeStdio serves the MCP JSON-RPC protocol over an LSP-style stdio stream.
func ServeStdio(r io.Reader, w io.Writer, s *Server) error {
	br := bufio.NewReader(r)
	for {
		body, err := readFramedMessage(br)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return writeFramedMessage(w, jsonRPCResponse{JSONRPC: "2.0", Error: &jsonRPCError{Code: -32700, Message: "parse error"}})
		}
		if req.ID == nil {
			continue
		}
		if err := writeFramedMessage(w, handleJSONRPCRequest(s, req)); err != nil {
			return err
		}
	}
}

func readFramedMessage(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			length, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length %q: %w", value, err)
			}
			contentLength = length
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func writeFramedMessage(w io.Writer, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func handleJSONRPCRequest(s *Server, req jsonRPCRequest) jsonRPCResponse {
	resp := jsonRPCResponse{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "plan-ai", "version": "2.0.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}
	case "tools/list":
		resp.Result = map[string]any{"tools": toolDefinitionsForProtocol(s.ListTools())}
	case "tools/call":
		name, _ := req.Params["name"].(string)
		args, _ := req.Params["arguments"].(map[string]any)
		result := s.ExecuteTool(name, args)
		textBytes, _ := json.MarshalIndent(result, "", "  ")
		resp.Result = map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(textBytes)}},
			"isError": !result.Success,
		}
	default:
		resp.Error = &jsonRPCError{Code: -32601, Message: "method not found"}
	}
	return resp
}

func toolDefinitionsForProtocol(tools []ToolDefinition) []map[string]any {
	out := make([]map[string]any, len(tools))
	for i, tool := range tools {
		out[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.Schema,
		}
	}
	return out
}
