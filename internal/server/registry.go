package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Registry struct {
	servers []*VirtualServer
	cfg     Config
}

func NewRegistry(cfg Config) *Registry {
	servers := make([]*VirtualServer, cfg.NumServers)
	for i := 0; i < cfg.NumServers; i++ {
		servers[i] = NewVirtualServer(i, cfg)
	}
	return &Registry{servers: servers, cfg: cfg}
}

func (reg *Registry) StartMutations(ctx context.Context) {
	for _, vs := range reg.servers {
		go vs.mutateLoop(ctx)
	}
	log.Printf("All %d servers initialized and mutation loops started", len(reg.servers))
}

func (reg *Registry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// path: /server/{id}/mcp
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "server" || parts[2] != "mcp" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 0 || id >= len(reg.servers) {
		http.Error(w, fmt.Sprintf("server %q not found", parts[1]), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodPost:
		handleMCP(reg.servers[id], w, r)
	case http.MethodGet:
		handleNotificationStream(reg.servers[id], w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
