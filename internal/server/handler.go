package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ernestkhasanzhinov/mcp-simulator/internal/jsonrpc"
)

const (
	mcpProtocolVersion = "2024-11-05"
	serverName         = "mcp-simulator"
	serverVersion      = "0.1.0"
)

func handleMCP(server *VirtualServer, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest,
			jsonrpc.NewErrorResponse(nil, jsonrpc.ParseError, "failed to read body"))
		return
	}

	var req jsonrpc.Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest,
			jsonrpc.NewErrorResponse(nil, jsonrpc.ParseError, "invalid JSON"))
		return
	}

	if req.JSONRPC != jsonrpc.Version {
		writeJSON(w, http.StatusBadRequest,
			jsonrpc.NewErrorResponse(req.ID, jsonrpc.InvalidRequest, "unsupported jsonrpc version"))
		return
	}

	switch req.Method {
	case "initialize":
		handleInitialize(server, req, w)
	case "notifications/initialized":
		w.WriteHeader(http.StatusAccepted)
	case "tools/list":
		handleToolsList(server, req, w)
	default:
		writeJSON(w, http.StatusOK,
			jsonrpc.NewErrorResponse(req.ID, jsonrpc.MethodNotFound,
				fmt.Sprintf("method %q not supported", req.Method)))
	}
}

func handleInitialize(server *VirtualServer, req jsonrpc.Request, w http.ResponseWriter) {
	result := map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    fmt.Sprintf("%s-%d", serverName, server.ID),
			"version": serverVersion,
		},
	}

	writeJSON(w, http.StatusOK, jsonrpc.NewResponse(req.ID, result))
}

func handleToolsList(server *VirtualServer, req jsonrpc.Request, w http.ResponseWriter) {
	result := map[string]interface{}{
		"tools": server.GetTools(),
	}
	writeJSON(w, http.StatusOK, jsonrpc.NewResponse(req.ID, result))
}

func handleNotificationStream(server *VirtualServer, w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	ch := server.Subscribe()
	defer server.Unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
