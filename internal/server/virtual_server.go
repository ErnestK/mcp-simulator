package server

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/ernestkhasanzhinov/mcp-simulator/internal/jsonrpc"
	"github.com/ernestkhasanzhinov/mcp-simulator/internal/tools"
)

type VirtualServer struct {
	ID            int
	tools         []tools.Tool
	toolsMu       sync.RWMutex
	notifications chan json.RawMessage
	cfg           Config
}

func NewVirtualServer(id int, cfg Config) *VirtualServer {
	return &VirtualServer{
		ID:            id,
		tools:         tools.GenerateRandom(cfg.MinTools, cfg.MaxTools),
		notifications: make(chan json.RawMessage, 16),
		cfg:           cfg,
	}
}

func (vs *VirtualServer) GetTools() []tools.Tool {
	vs.toolsMu.RLock()
	defer vs.toolsMu.RUnlock()
	result := make([]tools.Tool, len(vs.tools))
	copy(result, vs.tools)
	return result
}

func (vs *VirtualServer) ToolCount() int {
	vs.toolsMu.RLock()
	defer vs.toolsMu.RUnlock()
	return len(vs.tools)
}

func (vs *VirtualServer) notify(data json.RawMessage) {
	select {
	case vs.notifications <- data:
	default:
	}
}

func (vs *VirtualServer) mutateLoop(ctx context.Context) {
	for {
		interval := vs.cfg.MinMutationInterval +
			time.Duration(rand.Int63n(int64(vs.cfg.MaxMutationInterval-vs.cfg.MinMutationInterval)))

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}

		vs.toolsMu.Lock()
		newTools, event := tools.Mutate(vs.tools)
		vs.tools = newTools
		count := len(newTools)
		vs.toolsMu.Unlock()

		mutationType := "added"
		if event.Type == tools.MutationRemove {
			mutationType = "removed"
		} else if event.Type == tools.MutationUpdate {
			mutationType = "updated"
		}
		log.Printf("Server %d: %s tool %q (now %d tools)", vs.ID, mutationType, event.ToolName, count)

		notification := jsonrpc.NewNotification("notifications/tools/list_changed", nil)
		data, err := json.Marshal(notification)
		if err != nil {
			continue
		}
		vs.notify(data)
	}
}
